package tools

import (
	"context"
	"fmt"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/global"
	"net/http"
)

// 获取所有符合配置文件要求的 issue
func (i issueFunctions) GetAllMath() (issues map[string]*github.Issue, err error) {
	c := *global.Conf
	global.Sugar.Debugw("load workspace issues",
		"step", "start")
	opt := &github.IssueListByRepoOptions{}
	opt.State = "open"
	// 仅根据 kind/page 类型的 label 筛选 issue
	opt.Labels = []string{"kind/page"}

	// 每页 100 个 issue
	opt.Page = 1
	opt.PerPage = 100

	issues = make(map[string]*github.Issue)
	for {
		is, resp, err := global.Client.Issues.ListByRepo(
			context.TODO(),
			c.Repository.Spec.Workspace.Owner,
			c.Repository.Spec.Workspace.Repository,
			opt,
		)
		if err != nil {
			global.Sugar.Errorw("load issue list",
				"call api", "failed",
				"err", err.Error(),
			)
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			global.Sugar.Errorw("load issue list",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			return nil, err
		}

		for _, v := range is {
			// TODO 关闭重复 issue
			issues[v.GetTitle()] = v
		}

		if len(is) < opt.PerPage {
			break
		}
		opt.Page++
	}
	global.Sugar.Debugw("get all match issues",
		"len", len(issues))

	return issues, nil
}

// 获取 pr issue
func (i issueFunctions) GetPRIssue() *github.Issue {
	is, resp, err := global.Client.Issues.Get(context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		global.Conf.Repository.Spec.Workspace.Detection.PRIssue,
	)
	if err != nil {
		global.Sugar.Errorw("load commit issue",
			"call api", "failed",
			"err", err.Error(),
		)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		global.Sugar.Errorw("load commit issue",
			"call api", "unexpect status code",
			"status", resp.Status,
			"status code", resp.StatusCode,
			"response", resp.Body,
		)
		return nil
	}

	return is
}

func (i issueFunctions) Create(issue *github.IssueRequest) (newIssue *github.Issue, err error) {
	newIssue, resp, err := global.Client.Issues.Create(context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		issue,
	)
	if err != nil {
		global.Sugar.Errorw("create issues",
			"step", "call api",
			"title", issue.Title,
			"body", issue.Body,
			"err", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("create issues",
			"step", "check response code",
			"title", issue.Title,
			"body", issue.Body,
			"status code", resp.StatusCode,
			"resp body", string(body))
		err = fmt.Errorf("response code: %d, body:%s", resp.StatusCode, string(body))
		return
	}
	return
}

func (i issueFunctions) Edit(issue *github.Issue) (updatedIssue *github.Issue, err error) {
	return i.EditByIssueRequest(issue.GetNumber(), Convert.Issue(issue))
}

func (i issueFunctions) EditByIssueRequest(number int, issue *github.IssueRequest) (updatedIssue *github.Issue, err error) {
	if issue.Body == nil {
		global.Sugar.Errorw("edit issue",
			"confirm", "failed",
			"cause", "body can not be nil",
			"data", issue,
		)
		err = fmt.Errorf("body can not be nil")
		return
	}
	updatedIssue, resp, err := global.Client.Issues.Edit(context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		number,
		issue,
	)
	if err != nil {
		global.Sugar.Errorw("edit issues",
			"step", "call api",
			"number", number,
			"issue", issue,
			"err", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("edit issues",
			"step", "check response code",
			"number", number,
			"issue", issue,
			"status code", resp.StatusCode,
			"resp body", string(body))
		err = fmt.Errorf("response code: %d, body:%s", resp.StatusCode, string(body))
		return
	}
	return
}

// Comment
// 创建 issue comment
func (i issueFunctions) Comment(number int, body string) {
	// 如果 body 为空，则不做任何操作
	if body == "" {
		return
	}

	comment := &github.IssueComment{}
	comment.Body = &body
	_, resp, err := global.Client.Issues.CreateComment(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		number,
		comment)

	if err != nil {
		fmt.Printf("comment_issue_fail err: %v\n", err.Error())
		global.Sugar.Errorw("issue comment",
			"step", "call api",
			"status", "fail",
			"number", number,
			"body", body,
			"err", err.Error())
		return
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("issue comment",
			"step", "check response",
			"number", number,
			"status code", resp.StatusCode,
			"body", string(body))
		return
	}
}
