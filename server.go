// 主要通过 webhook 处理请求
package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"time"
)

func Start(t string, c int) {

	initClient(t)
	initOperator(c)

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

func initClient(token string) {
	// 初始化 client
	client = http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConnsPerHost:   5,
			MaxIdleConns:          5,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   60 * time.Second,
			ExpectContinueTimeout: 60 * time.Second,
		},
	}
	// 初始化 header
	header = http.Header{}
	header.Add("Accept", "application/vnd.github.v3+json")
	header.Add("Authorization", fmt.Sprintf("token %s", token))
}
