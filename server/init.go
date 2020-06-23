// init.go 对应 init 子命令的实现
// init 实现的是根据上游仓库和规则，在任务仓库创建初始化 issue
package server

import (
	"context"
	"fmt"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"issue-man/operation"
	"issue-man/tools"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 注意：这个 Init 并不是传统的初始化函数！
// Init 根据上游仓库的内容和 IssueCreate 规则，在工作仓库创建初始化 issue
// 获取 path 获取文件，
// 1. 根据规则（路径）获取全部上游文件
// 2. 根据规则（label）获取全部 issue，
// 3. 根据规则（路径），判断哪些 issue 需要新建
// 1. 包含 _index 开头的文件的目录，创建统一的 issue（但会继续遍历相关子目录），由 maintainer 统一管理。
// 3. 以包含 .md 文件的目录为单位，创建 issue（即一个目录可能包含多个 .md 文件）
func Init(conf config.Config) {
	fs, err := getUpstreamFiles()
	if err != nil {
		global.Sugar.Errorw("Get upstream files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}
	// init 始终基于最新 commit 来完成，
	// 所以这里直接更新 commit issue body
	commitIssue := operation.getPrIssue()
	commitIssue.Body = operation.genBodyBySha(operation.getLatestCommit())
	defer operation.updateCommitIssue(commitIssue)
	genAndCreateIssues(fs)
}

// 根据规则（路径）获取全部上游文件
func getUpstreamFiles() (files map[string]string, err error) {
	c := *global.Conf
	global.Sugar.Debugw("load upstream files",
		"step", "start")
	ts, resp, err := global.Client.Git.GetTree(context.TODO(),
		c.Repository.Spec.Source.Owner,
		c.Repository.Spec.Source.Repository,
		"master",
		true,
	)
	if err != nil {
		global.Sugar.Errorw("load upstream files",
			"call api", "Get tree",
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
		// 仅处理支持的文件类型
		if v.GetType() == "blob" && c.IssueCreate.SupportType(v.GetPath()) {
			files[v.GetPath()] = v.GetPath()
			continue
		}
	}
	//global.Sugar.Debugw("Get files",
	//	"data", files)
	return files, nil
}

// 根据规则（label）获取全部 issue，
// key 为 title
// title 由 parseTitleFromPath 生成
func getIssues() (issues map[string]*github.Issue, err error) {
	c := *global.Conf
	global.Sugar.Debugw("load workspace issues",
		"step", "start")
	opt := &github.IssueListByRepoOptions{}
	// 仅根据 kind 类型的 label 筛选 issue
	for _, v := range c.IssueCreate.Spec.Labels {
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
			issues[v.GetTitle()] = v
		}

		if len(is) < opt.PerPage {
			break
		}
		opt.Page++
	}
	global.Sugar.Debugw("Get issues",
		"data", issues)

	return issues, nil
}

// 根据配置、文件列表、已存在 issue，判断生成最终操作列表
// 遍历文件，判断文件是否符合条件，符合则直接创建
// （实现根据文件名生成 title、body。根据配置生成 label、assignees 的 issue）
// （实现根据 body 提取文件列表的方法）
// 1. 根据规则获取已存在 issue 列表
// 2. 遍历，根据 title 判断 issue 是否已经存在
// 3. 更新 issue（如果文件有变化），assignees 如果不为空，则不修改，如果为空则判断配置是否有配置 assignees，如都为空则不操作。
// 4. 创建 issue
func genAndCreateIssues(fs map[string]string) {
	conf := *global.Conf
	existIssues, err := getIssues()
	if err != nil {
		global.Sugar.Errorw("Get issues files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}
	// 更新和创建的 issue
	updates, creates := make(map[int]*github.IssueRequest), make(map[string]*github.IssueRequest)
	updateFail, createFail := 0, 0
	for file := range fs {
		for _, v := range conf.IssueCreate.Spec.Includes {
			// 符合条件的文件
			if global.Conf.IssueCreate.SupportFile(v, file) {
				// 根据 title 判断，如果已存在相关 issue，则更新
				exist := existIssues[*tools.Generate.Title(file)]
				if exist != nil {
					updates[*exist.Number] = updateIssue(false, file, *exist)
				} else {
					// 不存在，则新建，新建也分两种情况
					// 有多个新文件属于一个 issue
					create := creates[*tools.Generate.Title(file)]
					if create != nil {
						creates[*tools.Generate.Title(file)] = updateNewIssue(file, create)
					} else {
						// 是一个新的新 issue
						creates[*tools.Generate.Title(file)] = newIssue(v, file)
					}
				}
				// 文件已处理，break 内层循环
				break
			}
		}
	}

	//global.Sugar.Debugw("create issues",
	//	"data", creates)
	//global.Sugar.Debugw("update issues",
	//	"data", updates)

	wg := sync.WaitGroup{}
	lock := make(chan int, 5)
	go func() {
		// 将 API 频率限制为每秒 5 次
		for range lock {
			time.Sleep(time.Millisecond * 500)
		}
	}()
	// update
	for k, v := range updates {
		wg.Add(1)
		go func(number int, issue *github.IssueRequest) {
			defer wg.Done()
			lock <- 1
			_, resp, err := global.Client.Issues.Edit(
				context.TODO(),
				global.Conf.Repository.Spec.Workspace.Owner,
				global.Conf.Repository.Spec.Workspace.Repository,
				number,
				issue,
			)
			if err != nil {
				global.Sugar.Errorw("init issues",
					"step", "update",
					"id", number,
					"title", issue.Title,
					"body", issue.Body,
					"err", err.Error())
				updateFail++
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				global.Sugar.Errorw("init issues",
					"step", "update",
					"id", number,
					"title", issue.Title,
					"body", issue.Body,
					"status code", resp.StatusCode,
					"resp body", string(body))
				updateFail++
				return
			}
		}(k, v)
	}

	// create
	for _, v := range creates {
		wg.Add(1)
		go func(issue *github.IssueRequest) {
			defer wg.Done()
			lock <- 1
			_, resp, err := global.Client.Issues.Create(
				context.TODO(),
				global.Conf.Repository.Spec.Workspace.Owner,
				global.Conf.Repository.Spec.Workspace.Repository,
				issue,
			)
			if err != nil {
				global.Sugar.Errorw("init issues",
					"step", "create",
					"title", issue.Title,
					"body", issue.Body,
					"err", err.Error())
				createFail++
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				body, _ := ioutil.ReadAll(resp.Body)
				global.Sugar.Errorw("init issues",
					"step", "create",
					"title", issue.Title,
					"body", issue.Body,
					"status code", resp.StatusCode,
					"resp body", string(body))
				createFail++
				return
			}
		}(v)
	}
	wg.Wait()

	global.Sugar.Infow("init issues",
		"step", "done",
		"create", len(creates),
		"create fail", createFail,
		"update", len(updates),
		"update fail", updateFail,
	)
}

// 根据已存在的 issue 和配置，返回更新后的 issue
// 仅更新 body
func updateNewIssue(file string, exist *github.IssueRequest) *github.IssueRequest {
	exist.Body, _ = tools.Generate.Body(false, file, exist.GetBody())
	return exist
}

// 更新 issue request 的 body
func updateIssueRequest(remove bool, file string, exist *github.IssueRequest) *github.IssueRequest {
	exist.Body, _ = tools.Generate.Body(remove, file, exist.GetBody())
	return exist
}

// 根据已存在的 issue 和配置，返回更新后的 IssueRequest
func updateIssue(remove bool, file string, exist github.Issue) (update *github.IssueRequest) {
	length := 0
	update = &github.IssueRequest{}
	update.Title = exist.Title
	update.Body, length = tools.Generate.Body(remove, file, *exist.Body)

	// 对于已存在的 issue
	// label、assignees、milestone 不会变化
	if exist.Milestone != nil {
		update.Milestone = exist.Milestone.Number
	}
	if exist.Labels != nil {
		update.Labels = tools.Convert.Label(exist.Labels)
	}
	if exist.Assignees != nil {
		update.Assignees = tools.Convert.Assignees(exist.Assignees)
	}

	// 如果文件列表为 0，则添加需要检查的 Label
	// 反之则移除
	if length == 0 {
		update.Labels = tools.Convert.LabelAdd(update.Labels, global.Conf.Repository.Spec.Workspace.Detection.DeprecatedLabel...)
	} else {
		update.Labels = tools.Convert.LabelRemove(update.Labels, global.Conf.Repository.Spec.Workspace.Detection.DeprecatedLabel...)
	}
	return
}

func newIssue(include config.Include, file string) (new *github.IssueRequest) {
	new = &github.IssueRequest{}
	new.Title = tools.Generate.Title(file)
	new.Body, _ = tools.Generate.Body(false, file, "")

	new.Labels = tools.Convert.LabelAdd(&include.Labels, global.Conf.IssueCreate.Spec.Labels...)
	new.Assignees = tools.Get.String(global.Conf.IssueCreate.Spec.Assignees)
	new.Milestone = tools.Get.Int(global.Conf.IssueCreate.Spec.Milestone)
	return
}
