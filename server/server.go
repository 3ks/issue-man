package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/client"
	"issue-man/config"
	"log"
	"net/http"
)

var (
	// 仓库的完整名
	fullName string

	// 解析的事件列表
	// todo 配置化
	events []github.Event
)

func Start(token string) {
	conf, ok := viper.Get("config").(*config.Config)
	if !ok {
		panic("viper get config fail")
	}
	fullName = conf.FullRepositoryName

	// 初始化 GitHub client
	client.Init(token)

	// 初始化指令及 maintainer
	config.Init()

	// 初始化处理的事件列表
	events = []github.Event{github.IssueCommentEvent, github.PullRequestEvent}

	// 定义监听路由
	router := gin.Default()

	v1 := router.Group("/api/v1")
	v1.POST("/webhooks/", handler)
	v1.POST("/service-mesher/", handler)

	srv := &http.Server{
		Addr:    conf.Port,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

}
