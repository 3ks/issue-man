package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/global"
	"issue-man/operation"
	"net/http"
	"strconv"
	"time"
)

// issue-man 工作流程：
// 解析 webhook 数据：https://github.com/go-playground/webhooks
// 拼装数据：根据 GitHub API 要求，以及自身需要拼装数据
// 发送请求：https://github.com/google/go-github
func handler(c *gin.Context) {
	hook, _ := github.New()
	p, err := hook.Parse(c.Request, events...)
	if err != nil {
		fmt.Printf("unmarshal post payload err: %v", err.Error())
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

// 手动调用更新函数
func Sync(c *gin.Context) {
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
	syncIssues()
	c.JSON(http.StatusOK, gin.H{"status": "done"})
}

func issueComment(payload github.IssueCommentPayload) {
	// 不处理未知 repository 的事件
	if payload.Repository.FullName != global.Conf.Repository.Spec.Source.GetFullName() {
		return
	}

	// 不处理已关闭的 issue
	if payload.Issue.State == "closed" {
		return
	}

	is := ParseInstruct(payload.Comment.Body)
	// 未能解析出任何指令
	if len(is) == 0 {
		return
	}

	// 执行指令
	operation.IssueHanding(payload, is)
}

// 维护团队成员变化情况
func team(payload github.MembershipPayload) {
	if payload.Team.Name == global.Conf.Repository.Spec.Workspace.MaintainerTeam {
		global.LoadMaintainers()
	}
}

// 维护组织成员变化情况
func org(payload github.OrganizationPayload) {
	switch payload.Action {
	case "member_added", "member_removed":
		global.LoadMembers()
	}
}

// 有 pull request 事件
// 条件是 source 的仓库中有 pull request 被 close 且 merge 为 true，则触发检测方法，
// 检测是否有 issue 需要更新
func pr(payload github.PullRequestPayload) {
	if payload.Repository.FullName == global.Conf.Repository.Spec.Source.GetFullName() {
		if payload.Action == "closed" && payload.PullRequest.Merged {
			syncIssues()
		}
	}
}
