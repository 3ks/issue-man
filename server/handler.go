package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/operation"
	"net/http"
)

var repos map[string]bool
var events []github.Event

func init() {
	// todo 配置化
	// 仓库
	repos = make(map[string]bool)
	repos["gorda/gorda.io"] = true
	repos["servicemesher/istio-handbook"] = true

	// 事件
	events = []github.Event{github.IssueCommentEvent, github.PullRequestEvent}
}

// issue-man 会对事件做一些验证，例如：对应 issue 的状态，仓库是否在白名单。
// 通过验证后，issue-man 会向所有启用的 Operator 提供反序列化后的完整的事件 payload。
// Operator 通过解析 comment 行为，决定是否处理该事件。
// Operator 的具体行为交给实现了 operator 接口的对象。
// Operator 关注的重点在于如何正确提取并组装 request 的数据。
// 最后只需通过 client 包下的 Client 对象调用相关方法（接口）即可。
// 一般来说，在一次事件中，每一类指令会对应一个 Operator，每个 Operator 最终是否运行成功，与其它 Operator 无关。
//
// 简而言之：反序列化 webhook 发送过来的数据，判断并拼装数据（可能还需要额外的调用接口获取一些数据），调用相应的接口。
func handler(c *gin.Context) {

	p, err := hook.Parse(c.Request, events...)
	if err != nil {
		fmt.Printf("unmarshal post payload err: %v", err.Error())
	}

	switch p.(type) {
	case github.IssueCommentPayload:
		issueComment(p.(github.IssueCommentPayload))
	case github.PullRequestPayload:
	// todo
	default:
		// todo
	}

	c.String(http.StatusOK, "")
}

func issueComment(payload github.IssueCommentPayload) {
	// 不处理未知 repository 的事件
	if !repos[payload.Repository.FullName] {
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
	operation.IssueHanding(payload, is)
}
