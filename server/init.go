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
	"os"
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
		// 仅处理支持的文件类型
		if v.GetType() == "blob" && c.IssueCreate.SupportType(v.GetPath()) {
			files[v.GetPath()] = v.GetPath()
			continue
		}
	}
	//global.Sugar.Debugw("get files",
	//	"data", files)
	return files, nil
}

// 根据规则（label）获取全部 issue，
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
			issues[*v.Title] = v
		}

		if len(is) < opt.PerPage {
			break
		}
		opt.Page++
	}
	global.Sugar.Debugw("get issues",
		"data", issues)

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
	for file := range fs {
		for _, v := range conf.IssueCreate.Spec.Includes {
			// 符合条件的文件
			if global.Conf.IssueCreate.SupportFile(v, file) {
				// 根据 title 判断，如果已存在相关 issue，则更新
				exist := existIssues[*parseTitleFromPath(file)]
				if exist != nil {
					updates[*exist.Number] = updateIssue(false, file, *exist)
				} else {
					// 不存在，则新建，新建也分两种情况
					// 有多个新文件属于一个 issue
					create := creates[*parseTitleFromPath(file)]
					if create != nil {
						creates[*parseTitleFromPath(file)] = updateNewIssue(file, create)
					} else {
						// 是一个新的新 issue
						creates[*parseTitleFromPath(file)] = newIssue(v, file)
					}
				}
				// 文件已处理，break 内层循环
				break
			}
		}
	}

	//global.Sugar.Debugw("create issues",
	//	"data", creates)
	global.Sugar.Debugw("update issues",
		"data", updates)

	os.Exit(0)
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
			defer resp.Body.Close()
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
// 仅更新 body
func updateNewIssue(file string, exist *github.IssueRequest) *github.IssueRequest {
	exist.Body, _ = genBody(false, file, exist.GetBody())
	return exist
}

// 根据已存在的 issue 和配置，返回更新后的 issue
func updateIssue(remove bool, file string, exist github.Issue) (update *github.IssueRequest) {
	const (
		CHECK = "status/need-check"
	)

	length := 0
	update = &github.IssueRequest{}
	update.Title = exist.Title
	update.Body, length = genBody(remove, file, *exist.Body)

	// 对于已存在的 issue
	// label、assignees、milestone 不会变化
	update.Labels = convertLabel(exist.Labels)
	update.Assignees = convertAssignees(exist.Assignees)
	update.Milestone = exist.Milestone.Number

	// 如果文件列表为 0，则添加特殊 label
	// 反之则移除
	if length == 0 {
		tmp := append(*update.Labels, CHECK)
		update.Labels = &tmp
	} else {
		index := 0
		tmp := update.GetLabels()
		for _, v := range tmp {
			if v == CHECK {
				continue
			}
			(*update.Labels)[index] = v
			index++
		}
		tmp = tmp[:index]
		update.Labels = &tmp
	}
	return
}

func newIssue(include config.Include, file string) (new *github.IssueRequest) {
	new = &github.IssueRequest{}
	new.Title = parseTitleFromPath(file)
	new.Body, _ = genBody(false, file, "")

	// 创建新切片
	labels := append(*copySlice(global.Conf.IssueCreate.Spec.Labels), include.Labels...)
	new.Labels = &labels
	new.Assignees = copySlice(global.Conf.IssueCreate.Spec.Assignees)
	new.Milestone = copyInt(global.Conf.IssueCreate.Spec.Milestone)
	return
}

func copySlice(src []string) *[]string {
	dst := make([]string, len(src))
	copy(dst, src)
	return &dst
}

func copyInt(src int) *int {
	if src == 0 {
		return nil
	}
	return &src
}

// parseTitleFromPath 解析路径，生成 title
// 传入的路径总是这样的：content/en/faq/setup/k8s-migrating.md，预期 title 为： faq/setup
// 对于文件名为：_index 开头的文件，预期 title 总是为： Architecture
// 不会出现返回 nil 的情况，最差情况下返回值为 ""
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

// parseURLFormPath
// 根据 PATH 生成站点的  HTTPS URL
// TODO site 站点和 github 文件路径处理配置化
func parseURLFormPath(p string) (source, translate string) {
	// 去除两端路径
	t := strings.Split(strings.Replace(p, global.Conf.Repository.Spec.Source.RemovePrefix, "", 1), "/")
	tmp := path.Join(t[:len(t)-1]...)

	sourceSite := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(global.Conf.Repository.Spec.Source.Site, "/"), "https://"), "http://")
	translateSite := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(global.Conf.Repository.Spec.Translate.Site, "/"), "https://"), "http://")

	return fmt.Sprintf("https://%s", path.Join(sourceSite, tmp)), fmt.Sprintf("https://%s", path.Join(translateSite, tmp))
}

// genBody 根据文件名和旧的 body，生成新的 body
func genBody(remove bool, file, oldBody string) (body *string, length int) {
	t := ""
	body = &t
	oldBody = strings.ReplaceAll(oldBody, "\r\n", "\n")

	// map 存储去重
	files := make(map[string]string)
	files[file] = file
	lines := strings.Split(oldBody, "\n")
	for _, line := range lines {
		if strings.Contains(line, "content/en") { // TODO
			files[line] = line
		}
	}
	// 用于移除某个文件的情况
	if remove {
		delete(files, file)
	}

	length = len(files)
	fs := make([]string, len(files))
	// map 转 slice 以便排序
	count := 0
	for k := range files {
		fs[count] = k
		count++
	}
	// 排序
	sort.Slice(fs, func(i, j int) bool {
		return fs[i] < fs[j]
	})

	source, translate := parseURLFormPath(file)
	// 构造 body
	bf := bytes.Buffer{}
	// _index 类文件无统一页面
	if strings.Contains(file, "_index") {
		bf.WriteString(fmt.Sprintf("## Source\n\n#### Files\n\n"))
	} else {
		bf.WriteString(fmt.Sprintf("## Source\n\n#### URL\n\n%s\n\n#### Files\n\n", source))
	}
	for _, v := range fs {
		if v == "" {
			continue
		}
		// TODO 知识盲区
		bf.WriteString(fmt.Sprintf("- https://github.com/%s/%s/tree/master/%s\n\n",
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			v))
	}

	bf.WriteString("\n")

	// _index 类文件无统一页面
	if strings.Contains(file, "_index") {
		bf.WriteString(fmt.Sprintf("## Translate\n\n#### Files\n\n"))
	} else {
		bf.WriteString(fmt.Sprintf("## Translate\n\n#### URL\n\n%s\n\n#### Files\n\n", translate))
	}
	for _, v := range fs {
		if v == "" {
			continue
		}
		bf.WriteString(fmt.Sprintf("- https://github.com/%s/%s/tree/master/%s\n\n",
			global.Conf.Repository.Spec.Translate.Owner,
			global.Conf.Repository.Spec.Translate.Repository,
			strings.ReplaceAll(v, "content/en", "content/zh"))) // TODO
	}
	t = bf.String()
	return
}
