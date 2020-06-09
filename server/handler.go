package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/global"
	"issue-man/operation"
	"net/http"
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
	default:
	}

	c.String(http.StatusOK, "")
}

func issueComment(payload github.IssueCommentPayload) {
	// 不处理未知 repository 的事件
	if payload.Repository.FullName != global.Conf.Repository.Spec.Workspace.GetFullName() {
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
	if payload.Team.Name == *global.Conf.Repository.Spec.MaintainerTeam {
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
