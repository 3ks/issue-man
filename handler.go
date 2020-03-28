package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var repos map[string]bool

func init() {
	repos = make(map[string]bool)
	repos["gorda/gorda.io"] = true
	repos["servicemesher/istio-handbook"] = true
}

func handler(c *gin.Context) {
	payload := Payload{}

	err := c.Bind(&payload)
	if err != nil {
		fmt.Printf("unmarshal post payload err: %v", err.Error())
	}

	handlerPayload(payload)
	c.JSON(http.StatusOK, []string{})
}

// 判断 comment 目的
func handlerPayload(payload Payload) {

	i := payload.GetIssue()
	c := payload.GetComment()

	// 不处理未知 repository 的事件
	if !repos[i.urls.RepositoryName] {
		return
	}

	// 不处理已关闭的 issue
	if i.State == "closed" {
		return
	}

	// split \n ?
	// 有对应指令则执行对应操作
	for k, v := range ops {
		if strings.Contains(c.Body, k) {
			v(i, c)
			break
		}
	}
}
