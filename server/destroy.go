// destroy.go 对应 destroy 子命令的实现
// destroy 实现的是根据规则删除任务仓库的 issue。
package server

import (
	"context"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"net/http"
	"sync"
)

// Destroy 根据规则删除任务仓库的 issue。
func Destroy(conf config.Config) {
	issues, err := getIssues()
	if err != nil {
		global.Sugar.Errorw("get issues files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}
	wg := sync.WaitGroup{}

	for _, v := range issues {
		wg.Add(1)
		go func(issue *github.Issue) {
			defer wg.Done()
			issueRequest := issueToRequest(issue)
			state := "close"
			issueRequest.State = &state
			_, resp, err := global.Client.Issues.Edit(
				context.TODO(),
				global.Conf.Repository.Spec.Workspace.Owner,
				global.Conf.Repository.Spec.Workspace.Repository,
				*issue.Number,
				issueRequest,
			)
			if err != nil {
				global.Sugar.Errorw("destroy",
					"call api", "failed",
					"issue", *issue.Number,
					"err", err.Error(),
				)
				return
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				global.Sugar.Errorw("destroy",
					"call api", "unexpect status code",
					"issue", *issue.Number,
					"status", resp.Status,
					"status code", resp.StatusCode,
					"response", string(body),
				)
				return
			}
		}(v)
	}
	wg.Wait()

	global.Sugar.Infow("destroy issues",
		"step", "done")
}

// issueToRequest
// 将 github.Issue（API Response）转换为 github.IssueRequest（API Request）
func issueToRequest(issue *github.Issue) (ir *github.IssueRequest) {
	return &github.IssueRequest{
		Title:     issue.Title,
		Body:      issue.Body,
		State:     issue.State,
		Milestone: issue.Milestone.Number,
		Labels:    convertLabel(issue.Labels),
		Assignees: convertAssignees(issue.Assignees),
	}
}

// convertLabel
// 提取 github.Label 结构体内的 label名 字段
// 并组合为符合调用 API 要求的格式
func convertLabel(labels []*github.Label) *[]string {
	lb := make([]string, len(labels))
	for k, v := range labels {
		lb[k] = *v.Name
	}
	return &lb
}

// convertAssignees
// 提取 github.User 结构体内的用户名字段
// 并组合为符合调用 API 要求的格式
func convertAssignees(assignees []*github.User) *[]string {
	as := make([]string, len(assignees))
	for k, v := range assignees {
		as[k] = *v.Login
	}
	return &as
}
