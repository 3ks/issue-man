// 主要通过 webhook 处理请求
package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"issue-man/operator"
	"log"
	"net"
	"net/http"
	"time"
)

func Start(t string, c int) {

	initClient(t)
	operation.initOperator(c)

	router := gin.Default()

	v1 := router.Group("/api/v1")

	v1.POST("/webhooks/", handler)
	v1.POST("/service-mesher/", handler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 服务连接
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

}
