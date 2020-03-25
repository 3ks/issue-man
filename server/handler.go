package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"issue-man/model"
	"issue-man/operation"
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

// issue-man 会对事件做一些验证，例如：对应 issue 的状态，仓库是否在白名单。
// 通过验证后，issue-man 会向所有启用的 Operator 提供两方面的数据：request.Header 和事件 payload。其中：
//		request.Header 包含了鉴权所需的内容，在请求时，可以带上。
// 		事件 payload 包含了事件完整的数据。
// Operator 一般解析 comment 行为，决定是否处理该事件。
// Operator 的具体行为交给实现了 operator 接口的对象。
// Operator 最终都会以某种方法调用某个 GitHub API V3 的接口。其中：
// 		大多数情况下，接口的具体 URL 都可以在 `事件 payload` 中找到，而不需要自行拼接。
// Operator 关注的重点在于如何正确组装 request body 的数据。
//
// 一般来说，在一次事件中，每一类指令会对应一个 Operator，每个 Operator 最终是否运行成功，与其它 Operator 无关。
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
	operation.Handing(i,c)
}
