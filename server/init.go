// init.go 对应 init 子命令的实现
// init 实现的是根据上游仓库和规则，在任务仓库创建初始化 issue
package server

import (
	"context"
	"fmt"
	"github.com/google/go-github/v30/github"
	"issue-man/config"
	"issue-man/global"
	"net/http"
	"path"
	"strings"
)

// 注意：这个 Init 并不是传统的初始化函数！
// Init 根据上游仓库和 IssueCreate 规则，在任务仓库创建初始化 issue
// 获取 path 获取文件，
// 1. 根据规则（路径）获取全部上游文件
// 2. 根据规则（label）获取全部 issue，
// 3. 根据规则（路径），判断哪些 issue 需要新建
// 1. 包含 _index 开头的文件的目录，创建统一的 issue（但会继续遍历相关子目录），由 maintainer 统一管理。
// 3. 以包含 .md 文件的目录为单位，创建 issue（即一个目录可能包含多个 .md 文件）
func Init(conf config.Config) {
	fs, err := getUpstreamFiles(conf)
	if err != nil {
		global.Sugar.Errorw("get upstream files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}

	//issues, err := getIssues(conf)
	//if err != nil {
	//	global.Sugar.Errorw("get issues files",
	//		"status", "fail",
	//		"err", err.Error(),
	//	)
	//	return
	//}

	genAndCreateIssues(conf, fs)
}

// 根据规则（路径）获取全部上游文件
func getUpstreamFiles(c config.Config) (files map[string]string, err error) {
	global.Sugar.Debugw("load upstream files list",
		"step", "start")
	ts, resp, err := global.Client.Git.GetTree(context.TODO(),
		c.Repository.Spec.Upstream.Owner,
		c.Repository.Spec.Upstream.Repository,
		"master",
		true,
	)
	if err != nil {
		global.Sugar.Errorw("load upstream files list",
			"call api", "get tree",
			"err", err.Error(),
		)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		global.Sugar.Errorw("load upstream files list",
			"call api", "unexpect status code",
			"status", resp.Status,
			"status code", resp.StatusCode,
			"response", resp.Body,
		)
		return nil, fmt.Errorf("%v\n", resp.Body)
	}
	files = make(map[string]string)
	for _, v := range ts.Entries {
		// 目标目录下的文件
		if strings.HasPrefix(*v.Type, *c.IssueCreate.Spec.Content) && *v.Type == "blob" {
			// 仅处理 md 和 html 文件
			if path.Ext(*v.Path) == ".md" || path.Ext(*v.Path) == ".html" {
				// key path
				// value path
				files[*v.Path] = *v.Path
			}
		}
	}
	return files, nil
}

// 根据规则（label）获取全部 issue，
func getIssues(c config.Config) (issues map[string]*github.Issue, err error) {
	global.Sugar.Debugw("load upstream files list",
		"step", "start")
	opt := &github.IssueListByRepoOptions{}
	// 仅根据 kind 类型的 label 筛选 issue
	for _, v := range *c.IssueCreate.Spec.Labels {
		if strings.HasPrefix(v, "kind/") {
			opt.Labels = append(opt.Labels, v)
		}
	}
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
			issues[*v.Title] = v
		}

		if len(is) < opt.PerPage {
			break
		}
		opt.Page++
	}

	return issues, nil
}

// TODO 定时检测应该存在的 issue，和实际存在的 issue
// TODO 并做出添加、更新、删除操作。（频率较低）
// TODO 根据 commit 做出操作（频率较高）
// 根据配置、文件列表、已存在 issue，判断生成最终操作列表
// 遍历文件，判断文件是否符合条件，符合则直接创建
// （实现根据文件名生成 title、body。根据配置生成 label、assignees 的 issue）
// （实现根据 body 提取文件列表的方法）
// 1. 根据规则获取已存在 issue 列表
// 2. 遍历，根据 title 判断 issue 是否已经存在
// 3. 更新 issue（如果文件有变化），assignees 如果不为空，则不修改，如果为空则判断配置是否有配置 assignees，如都为空则不操作。
// 4. 创建 issue
func genAndCreateIssues(conf config.Config, fs map[string]string) {
	issues := make(map[string]*github.IssueRequest)
	for k := range fs {
		for _, v := range conf.IssueCreate.Spec.Includes {
			// 符合条件
			if v.OK(k) {
				title := ""
				issue := &github.IssueRequest{
					Title:     &fs[k],
					Body:      nil,
					Labels:    nil,
					Assignee:  nil,
					Milestone: nil,
					Assignees: nil,
				}
				// _index 文件汇聚
				if *v.Path == "_index" {

				} else {

				}
				// break 内层循环
				break
			}
			//
		}
	}
}

const (
	Create = "create"
	Update = "update"
	Delete = "delete"
)

// 判断文件是否
func matchFile(rule config.IssueCreate, file string) (op string) {

	return ""
}

func newIssue() *github.IssueRequest {
	return &github.IssueRequest{
		Title:     nil,
		Body:      nil,
		Labels:    global.Conf.IssueCreate.Spec.Labels,
		Assignees: global.Conf.IssueCreate.Spec.Assignees,
		Milestone: global.Conf.IssueCreate.Spec.Milestone,
	}
}
