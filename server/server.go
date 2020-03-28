// 主要通过 webhook 处理请求
package server

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/webhooks.v5/github"
	"log"
	"net/http"
)

var (
	hook *github.Webhook
)
func Start(t string, c int) {

	//initClient(t)
	//operation.initOperator(c)

	hook, _ = github.New()

	router := gin.Default()

	v1 := router.Group("/api/v1")

	v1.POST("/webhooks/", handler)
	v1.POST("/service-mesher/", handler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}


	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

}
