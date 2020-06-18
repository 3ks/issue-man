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
	"log"
	"net/http"
)

var (
	// 解析的事件列表
	events []github.Event
)

// TODO 定期检查同步状态。
func Start(conf config.Config) {
	// 初始化处理的事件列表
	events = []github.Event{
		github.IssueCommentEvent,
		github.MembershipEvent,
		github.OrganizationEvent,
	}

	// 定时检测任务
	go job()

	// 定义监听路由
	router := gin.Default()

	v1 := router.Group("/api/v1")
	v1.POST("/webhooks/", handler)
	v1.GET("/sync", Sync)

	srv := &http.Server{
		Addr:    global.Conf.Repository.Spec.Port,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
