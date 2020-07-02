package comm

import (
	"github.com/google/uuid"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/global"
)

// 存储 IssueCommentPayload 里的一些信息
// 基本是目前进行各种操作需要用到的信息
type Info struct {
	// 仓库信息
	Owner      string
	Repository string

	// 评论人信息
	Login string
	// 评论提及到的人
	Mention []string

	// Issue 目前的信息
	IssueURL    string
	IssueNumber int
	Title       string
	Body        string
	Milestone   int
	State       string
	Assignees   []string
	Labels      []string

	// 一个指令的 UUID
	ReqID string
}

// Info
// 从 IssueCommentPayload 里的一些信息
// 避免多次书写出现错误
func (p *Info) Parse(payload github.IssueCommentPayload) {
	defer func() {
		if pc := recover(); pc != nil {
			global.Sugar.Errorw("Info pc",
				"req_id", p.ReqID,
				"pc", pc)
		}
	}()

	p.ReqID = uuid.New().String()
	p.Owner = payload.Repository.Owner.Login
	p.Repository = payload.Repository.Name

	p.Login = payload.Sender.Login

	p.IssueURL = payload.Issue.URL
	p.IssueNumber = int(payload.Issue.Number)
	p.Title = payload.Issue.Title
	p.Body = payload.Issue.Body
	p.State = payload.Issue.State

	if payload.Issue.Milestone != nil {
		p.Milestone = int(payload.Issue.Milestone.Number)
	}

	p.Assignees = make([]string, len(payload.Issue.Assignees))
	p.Labels = make([]string, len(payload.Issue.Labels))
	for i := 0; i < len(payload.Issue.Assignees) || i < len(payload.Issue.Labels); i++ {
		if i < len(p.Assignees) {
			p.Assignees[i] = payload.Issue.Assignees[i].Login
		}
		if i < len(p.Labels) {
			p.Labels[i] = payload.Issue.Labels[i].Name
		}
	}
}
