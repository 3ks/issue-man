package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"issue-man/model"
	"issue-man/operator"
	"net/http"
)

var repos map[string]bool

func init() {
	// todo 配置化
	repos = make(map[string]bool)
	repos["gorda/gorda.io"] = true
	repos["servicemesher/istio-handbook"] = true
}

func handler(c *gin.Context) {
	payload := model.IssueHook{}

	err := c.Bind(&payload)
	if err != nil {
		fmt.Printf("unmarshal post payload err: %v", err.Error())
	}

	handlerPayload(payload)
	c.JSON(http.StatusOK, []string{})
}

// 判断 comment 目的
func handlerPayload(payload model.IssueHook) {

	i := payload.GetIssue()
	c := payload.GetComment()

	// 不处理未知 repository 的事件
	if !repos[i.URL.RepositoryName] {
		return
	}

	// 不处理已关闭的 issue
	if i.State == "closed" {
		return
	}
	operator.Handing(i,c)
}
