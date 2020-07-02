// start.go 对应 start 子命令的实现
// start 实现的是：
// 1. 启动 HTTP 服务，监听 Webhook 事件，响应任务仓库的指令。
// 2. 定时检测上游仓库的更新，分析操作，根据规则对任务仓库的 issue 做出处理（新增，更新通知，删除）
package server

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/config"
	"issue-man/global"
	"issue-man/operation"
	"issue-man/tools"
	"log"
	"net/http"
	"strconv"
	"time"
)

// TODO 定期检查同步状态。
func Start(conf config.Config) {
	// 定时检测任务
	go operation.Sync()

	// 初始化处理的事件列表

	// 定义监听路由
	router := gin.Default()
	if global.Conf.Repository.Spec.LogLevel != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	v1 := router.Group("/api/v1")
	{
		v1.GET("/sync", check, Sync)
		v1.GET("/load", check, Load)
		v1.POST("/webhooks/", Webhooks)
	}

	// TODO 获取 project card 列表
	srv := &http.Server{
		Addr:    global.Conf.Repository.Spec.Port,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// 假冒伪劣中间件
func check(c *gin.Context) {
	if global.Conf.Repository.Spec.LogLevel == "dev" {
		c.Next()
		return
	}
	// 假装要一个 token
	reqUnix, err := strconv.Atoi(c.Query("token"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "unauthorized"})
		return
	}
	sub := time.Now().Unix() - int64(reqUnix)
	// 只允许 10 秒钟的偏差
	if sub > 10 || sub < -10 {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "unauthorized"})
		return
	}
	c.Next()
}

// 手动调用更新函数
func Sync(c *gin.Context) {
	operation.SyncIssues()
	c.JSON(http.StatusOK, gin.H{"status": "done"})
}

// 更新 maintainer 和 member 列表
func Load(c *gin.Context) {
	global.LoadMembers()
	global.LoadMaintainers()
	c.JSON(http.StatusOK, gin.H{"status": "done"})
}

// issue-man 工作流程：
// 解析 webhook 数据：https://github.com/go-playground/webhooks
// 拼装数据：根据 GitHub API 要求，以及自身需要拼装数据
// 发送请求：https://github.com/google/go-github
func Webhooks(c *gin.Context) {
	hook, _ := github.New()
	// 解析的事件列表
	p, err := hook.Parse(c.Request,
		github.IssueCommentEvent,
		github.MembershipEvent,
		github.OrganizationEvent,
		github.PullRequestEvent,
	)
	if err != nil {
		global.Sugar.Errorw("unmarshal payload",
			"status", "fail",
			"err", err.Error())
	}
	switch p.(type) {
	case github.IssueCommentPayload:
		issueComment(p.(github.IssueCommentPayload))
	case github.OrganizationPayload:
		org(p.(github.OrganizationPayload))
	case github.MembershipPayload:
		team(p.(github.MembershipPayload))
	case github.PullRequestPayload:
		pr(p.(github.PullRequestPayload))
	default:
	}

	c.JSON(http.StatusOK, nil)
}

// issueComment
// webhook payload 数据是 issue comment 事件
func issueComment(payload github.IssueCommentPayload) {
	// 只处理 workspace 组织的 comment 事件
	if payload.Repository.Owner.Login != tools.Get.WorkspaceOwner() {
		return
	}

	// 不处理已关闭的 issue
	if payload.Issue.State == "closed" {
		return
	}

	is := tools.Parse.Instruct(payload.Comment.Body)
	// 未能解析出任何指令
	if len(is) == 0 {
		return
	}

	// 执行指令
	operation.IssueHanding(payload, is)
}

// org
// webhook payload 数据是 org 事件
// 维护 workspace 组织成员变化情况
func org(payload github.OrganizationPayload) {
	// 只处理 workspace 组织的 organization 的事件
	if payload.Organization.Login != tools.Get.WorkspaceOwner() {
		return
	}
	switch payload.Action {
	case "member_added", "member_removed":
		global.LoadMembers()
	}
}

// team
// webhook payload 数据是 team 事件
// 维护 workspace 组织 maintainer team 成员的变化情况
func team(payload github.MembershipPayload) {
	// 只处理 workspace 组织的事件
	if payload.Organization.Login != tools.Get.WorkspaceOwner() {
		return
	}

	// 只处理 maintainer team 的事件
	if payload.Team.Name == global.Conf.Repository.Spec.Workspace.MaintainerTeam {
		global.LoadMaintainers()
	}
}

// pr
// webhook payload 数据是 pull request 事件
// 条件是 source 的仓库中有 pull request 被 close 且 merge 为 true，则触发检测方法，
// 检测是否有 issue 需要更新
func pr(payload github.PullRequestPayload) {
	// 只处理 source 仓库的 merged pr 事件
	if payload.Repository.FullName == global.Conf.Repository.Spec.Source.GetFullName() {
		if payload.Action == "closed" && payload.PullRequest.Merged {
			operation.SyncIssues()
		}
	}
}
