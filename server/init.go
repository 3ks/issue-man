// init.go 对应 init 子命令的实现
// init 实现的是根据上游仓库和规则，在任务仓库创建初始化 issue
package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
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
	fs, err := getUpstreamFiles()
	if err != nil {
		global.Sugar.Errorw("get upstream files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}

	genAndCreateIssues(fs)
}

// 根据规则（路径）获取全部上游文件
func getUpstreamFiles() (files map[string]string, err error) {
	c := *global.Conf
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
func getIssues() (issues map[string]*github.Issue, err error) {
	c := *global.Conf
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
func genAndCreateIssues(fs map[string]string) {
	conf := *global.Conf
	existIssues, err := getIssues()
	if err != nil {
		global.Sugar.Errorw("get issues files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}
	// 更新和创建的 issue
	updates, creates := make(map[int]*github.IssueRequest), make(map[string]*github.IssueRequest)
	for k := range fs {
		for _, v := range conf.IssueCreate.Spec.Includes {
			// 符合条件的文件
			if v.OK(k) {
				// 根据 title 判断，如果已存在，则更新
				if exist := existIssues[*parseTitleFromPath(k)]; exist != nil {
					updates[*exist.Number] = updateIssue(*v, k, *exist)
				} else {
					// 不存在，则新建
					creates[k] = newIssue(*v, k)
				}
				// 文件已处理，break 内层循环
				break
			}
		}
	}

	wg := sync.WaitGroup{}
	// update
	for k, v := range updates {
		wg.Add(1)
		go func(number int, issue *github.IssueRequest) {
			defer wg.Done()
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
				return
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := ioutil.ReadAll(resp.Body)
				global.Sugar.Errorw("init issues",
					"step", "update",
					"id", number,
					"title", issue.Title,
					"body", issue.Body,
					"status code", resp.StatusCode,
					"resp body", string(body))
				return
			}
		}(k, v)
	}

	// create
	for _, v := range creates {
		wg.Add(1)
		go func(issue *github.IssueRequest) {
			defer wg.Done()
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
				return
			}
			if resp.StatusCode != http.StatusCreated {
				body, _ := ioutil.ReadAll(resp.Body)
				global.Sugar.Errorw("init issues",
					"step", "create",
					"title", issue.Title,
					"body", issue.Body,
					"status code", resp.StatusCode,
					"resp body", string(body))
				return
			}
		}(v)
	}
	wg.Wait()

	global.Sugar.Infow("init issues",
		"step", "done")
}

// 根据已存在的 issue 和配置，返回更新后的 issue
func updateIssue(include config.Include, file string, exist github.Issue) (new *github.IssueRequest) {
	new = &github.IssueRequest{}
	new.Title = exist.Title
	new.Body = genBody(file, *exist.Body)

	// 对于已存在的 issue
	// label、assignees、milestone 不会变化
	labels, assignees := make([]string, 0), make([]string, 0)
	for _, v := range exist.Labels {
		labels = append(labels, *v.Name)
	}
	for _, v := range exist.Assignees {
		assignees = append(assignees, *v.Login)
	}
	new.Labels = &labels
	new.Assignees = &assignees
	new.Milestone = exist.Milestone.Number
	return
}

func newIssue(include config.Include, file string) (new *github.IssueRequest) {
	new = &github.IssueRequest{}
	new.Title = parseTitleFromPath(file)
	new.Body = genBody(file, "")

	labels := *global.Conf.IssueCreate.Spec.Labels
	for _, v := range include.Labels {
		labels = append(labels, *v)
	}

	new.Labels = &labels
	new.Assignees = global.Conf.IssueCreate.Spec.Assignees
	new.Milestone = global.Conf.IssueCreate.Spec.Milestone
	return
}

// parseTitleFromPath 解析路径，生成 title
// 传入的路径总是这样的：content/en/faq/setup/k8s-migrating.md，预期 title 为： faq/setup
// 对于文件名为：_index 开头的文件，预期 title 总是为： Architecture
func parseTitleFromPath(p string) (title *string) {
	tmp := ""
	title = &tmp
	if strings.ReplaceAll(path.Base(p), path.Ext(p), "") == "_index" {
		tmp = "Architecture"
		return
	}
	p = strings.Replace(p, "content/en/", "", 1)
	t := strings.Split(p, "/")
	tmp = strings.Join(t[:len(t)-1], "/")
	return
}

// genBody 根据文件名和旧的 body，生成新的 body
func genBody(file, oldBody string) (body *string) {
	t := ""
	body = &t

	// map 存储去重
	files := make(map[string]string)
	files[file] = file
	for _, v := range strings.Split(oldBody, "\n") {
		if strings.Contains(v, "content/en/") {
			files[v] = v
		}
	}
	// map 转 slice 以便排序
	fs := make([]string, len(files))
	count := 0
	for k := range files {
		fs[count] = k
		count++
	}
	// 排序
	sort.Slice(fs, func(i, j int) bool {
		return fs[i] < fs[j]
	})

	// 构造 body
	bf := bytes.Buffer{}
	bf.WriteString("# EN Files\n\n")
	for _, v := range fs {
		bf.WriteString(fmt.Sprintf("- https://github.com/istio/istio.io/tree/master/%s\n", v))
	}

	bf.WriteString("# ZH Files\n\n")
	for _, v := range fs {
		bf.WriteString(fmt.Sprintf("- https://github.com/istio/istio.io/tree/master/%s\n", strings.ReplaceAll(v, "content/en", "content/zh")))
	}
	t = bf.String()
	return
}
