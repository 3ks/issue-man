package server

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/google/go-github/v30/github"
	"io"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"net/http"
	"strings"
	"sync"
	"time"
)

// job
// 目前主要完成状态持续时间的检测，并提醒
// 思路：对于需要检测的状态（label），会将其添加至相应的切片
//      每天定时检测，满足相关条件时，则执行一些操作
//
// TODO 检测频率
// 1. 获取所有特定 label 的 issue
// 2. 获取存储 commit 的 issue
// 3. 遍历 commit，存储到栈内，直至第二步匹配的 commit。
// 4. pop commit 栈，分析涉及的文件，是否存在匹配的 issue
// 5. 对匹配的 issue，comment 提示，该 issue 对应的某个文件在哪次 commit 有变动
func job() {
	global.Sugar.Info("loaded jobs", "list", global.Jobs)
	// 解析检测时间
	t, err := time.ParseInLocation("2006-01-02 15:04",
		time.Now().Format("2006-01-02 ")+global.Conf.Repository.Spec.Workspace.DetectionAt,
		time.Local)
	if err != nil {
		global.Sugar.Errorw("parse detection time",
			"status", "fail")
		return
	}

	// 首次检测等待时间
	var s time.Duration
	// 今天已过，则等明天的这个时刻
	if t.Unix() <= time.Now().Unix() {
		s = t.AddDate(0, 0, 1).Sub(time.Now())
	} else {
		// 否则，等待今天的这个时刻
		s = t.Sub(time.Now())
	}
	global.Sugar.Info("waiting for to detection",
		"sleep", s.String())
	time.Sleep(s)

	for {
		// 同步检测是一个特殊的任务，会检测两次 commit 之间所有 commit 涉及的文件，并提示
		syncIssues()

		// 遍历检测任务
		//for _, v := range global.Jobs {
		//operation.Job(conf.FullRepositoryName, v)
		//}

		// 等待明天的这个时刻
		t = t.AddDate(0, 0, 1)
		s = t.Sub(time.Now())
		global.Sugar.Info("waiting for to detection",
			"sleep", s.String())
		time.Sleep(s)
	}
}

//
func syncIssues() {
	commitIssue := getCommitIssue()
	if commitIssue == nil {
		return
	}
	// 获取 commit 列表
	commits := getRangeCommits(getShaFromCommitIssue(*commitIssue.Body))
	if len(commits) == 0 {
		global.Sugar.Warnw("get issues files",
			"status", "abnormal",
			"length", "0",
		)
		return
	}
	// 最近一次 commit
	commitIssue.Body = genBodyBySha(commits[0])
	// 更新 commit issue
	defer updateCommitIssue(commitIssue)

	// 获取 pr 列表
	prs := getAssociatedPRs(commits)

	// 获取 pr 涉及的文件列表
	files := getAssociatedFiles(prs)

	// 获取现有 issue 列表
	existIssues, err := getIssues()
	if err != nil {
		global.Sugar.Errorw("get issues files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}

	wg := sync.WaitGroup{}
	// 遍历文件
	// 判断是否匹配
	// 做出不同操作
	lock := make(chan int, 5)
	go func() {
		// 将 API 频率限制为每秒 2 次
		for range lock {
			time.Sleep(time.Millisecond * 500)
		}
	}()
	for _, file := range files {
		wg.Add(1)
		go func(file File) {
			defer wg.Done()
			// 1. 判断是否需要处理
			for _, include := range global.Conf.IssueCreate.Spec.Includes {
				// 符合条件的文件
				if global.Conf.IssueCreate.SupportFile(include, file.cf.GetFilename()) {
					lock <- 1
					file.Sync(
						include,
						existIssues[*parseTitleFromPath(file.cf.GetFilename())],
						existIssues[*parseTitleFromPath(file.cf.GetPreviousFilename())],
					)
					// 文件已处理
					return
				}
			}
		}(file)
	}
	wg.Wait()
}

func getShaFromCommitIssue(body string) string {
	tmp := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	if len(tmp) == 0 {
		return ""
	}
	return tmp[0]
}

// 生成带链接的 commit issue
func genBodyBySha(sha string) *string {
	body := fmt.Sprintf("%s\n\nhttps://github.com/%s/%s/tree/%s",
		sha,
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		sha,
	)
	return &body
}

type File struct {
	PrNumber int `json:"pr_number"`
	cf       *github.CommitFile
}

// 处理需同步文件
func (f File) Sync(include config.Include, existIssue, preIssue *github.Issue) {
	// 这里的操作指的是文件的操作
	// 至于 issue 是否存在，调用何种方法，需要额外判断
	const (
		ADD    = "added"
		MODIFY = "modified"
		RENAME = "renamed"
		REMOVE = "removed"
	)
	switch *f.cf.Status {
	// 更新 issue，不存在则创建 issue
	case ADD, MODIFY:
		// 更新 issue
		if existIssue != nil {
			f.update(existIssue)
		} else {
			// 创建 issue
			f.create(include)
		}
	// 重命名/移动文件
	case RENAME:
		f.rename(include, existIssue, preIssue)
	// 移除文件
	case REMOVE:
		f.remove(existIssue)
	default:
		global.Sugar.Warnw("unknown status",
			"file", f,
			"status", *f.cf.Status)
	}
}

// 更新 issue，并 comment 如果 issue 不存在，则创建
func (f File) create(include config.Include) {
	issue := newIssue(include, *f.cf.Filename)
	_, resp, err := global.Client.Issues.Create(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		issue,
	)
	if err != nil {
		global.Sugar.Errorw("sync create issues",
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
}

// 更新 issue，并 comment
func (f File) update(existIssue *github.Issue) (*github.Issue, error) {
	// 更新
	updatedIssue, err := f.edit(
		updateIssue(true, f.cf.GetPreviousFilename(), *existIssue),
		existIssue.GetNumber(),
		"update",
	)
	if err != nil {
		return nil, err
	}

	// comment
	err = f.comment(existIssue)
	if err != nil {
		return nil, err
	}
	return updatedIssue, nil
}

func (f File) comment(issue *github.Issue) error {
	// TODO 传入操作类型，对于 removed，仅 comment，不 @assigner
	// comment
	body := ""
	bf := bytes.Buffer{}
	bf.WriteString("maintainer: ")
	for _, v := range issue.Assignees {
		bf.WriteString(fmt.Sprintf("@%s ", v.GetLogin()))
	}
	bf.WriteString(fmt.Sprintf("\nstatus: %s", f.cf.GetStatus()))
	bf.WriteString(fmt.Sprintf("\npr: https://github.com/istio/istio.io/pull/%d", f.PrNumber))
	bf.WriteString(fmt.Sprintf("\ndiff: https://github.com/istio/istio.io/pull/%d/files#diff-%s",
		f.PrNumber, f.getFileHash()))
	body = bf.String()

	comment := &github.IssueComment{}
	comment.Body = &body
	_, resp, err := global.Client.Issues.CreateComment(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		issue.GetNumber(),
		comment)

	if err != nil {
		global.Sugar.Errorw("sync issue comment",
			"step", "call api",
			"status", "fail",
			"file", f,
			"err", err.Error())
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("CheckCount",
			"step", "parse response",
			"status", "fail",
			"statusCode", resp.StatusCode,
			"body", string(body))
		return err
	}
	return nil
}

func (f File) getFileHash() string {
	hash := md5.New()
	_, _ = io.WriteString(hash, f.cf.GetFilename())
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// 删除 issue 中的文件，更新后 issue 内无文件的情况，添加特殊 label 标识，maintainer 手动处理
func (f File) remove(preIssue *github.Issue) {
	if preIssue == nil {
		global.Sugar.Warnw("remove exist file issue",
			"status", "has no match issue",
			"file", f)
		return
	}
	f.edit(
		updateIssue(true, f.cf.GetPreviousFilename(), *preIssue),
		preIssue.GetNumber(),
		"remove",
	)
}

// 对于 renamed 文件，需要：
// 1. 更新/创建 新的 issue
// 2. 在旧的 issue 中移除对应的文件
func (f File) rename(include config.Include, existIssue, preIssue *github.Issue) {
	// 更新 issue
	if existIssue != nil {
		// preIssue 为空，则仅更新 existIssue
		// 这种极端情况很难出现
		if preIssue == nil {
			f.update(existIssue)
			global.Sugar.Warnw("renamed file issue",
				"status", "has no match previous issue",
				"file", f)
			return
		}
		// existIssue 和 preIssue 是同一个 issue
		if existIssue.GetNumber() == preIssue.GetNumber() {
			// 由于是同一个 issue，可以一次性完成更新，移除
			updatedIssue, err := f.edit(
				updateIssueRequest(true, f.cf.GetPreviousFilename(), updateIssue(false, *f.cf.Filename, *existIssue)),
				existIssue.GetNumber(),
				"renamed",
			)
			if err != nil {
				return
			}
			// comment
			f.comment(updatedIssue)
		} else {
			// 由于 existIssue 和 preIssue 不是同一个 issue
			// 需要分别完成更新、移除
			f.update(existIssue)
			f.remove(preIssue)
		}
	} else {
		// existIssue == nil，
		// 此时，创建 issue，并在 preIssue 中移除旧文件名
		f.create(include)
		// 尝试移除
		if preIssue != nil {
			f.remove(preIssue)
		}
	}
}

func (f File) edit(issue *github.IssueRequest, number int, option string) (*github.Issue, error) {
	updatedIssue, resp, err := global.Client.Issues.Edit(
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
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("edit issues",
			"step", option,
			"id", number,
			"title", issue.Title,
			"body", issue.Body,
			"status code", resp.StatusCode,
			"resp body", string(body))
		return nil, err
	}
	return updatedIssue, nil
}

func getAssociatedFiles(prs []int) []File {
	files := make([]File, 0)

	for _, v := range prs {
		for {
			opt := &github.ListOptions{
				Page:    1,
				PerPage: 3000,
			}
			tmp, resp, err := global.Client.PullRequests.ListFiles(
				context.TODO(),
				global.Conf.Repository.Spec.Source.Owner,
				global.Conf.Repository.Spec.Source.Repository,
				v,
				opt)
			if err != nil {
				global.Sugar.Errorw("load pr file list",
					"call api", "failed",
					"err", err.Error(),
				)
				return nil
			}
			if resp.StatusCode != http.StatusOK {
				global.Sugar.Errorw("load pr file list",
					"call api", "unexpect status code",
					"status", resp.Status,
					"status code", resp.StatusCode,
					"response", resp.Body,
				)
				return nil
			}
			for _, cf := range tmp {
				files = append(files, File{
					PrNumber: v,
					cf:       cf,
				})
			}
			// 结束内层循环
			if len(tmp) < opt.PerPage {
				break
			}
			opt.Page++
		}
	}
	return files
}

func getAssociatedPRs(commits []string) []int {
	prs := make([]int, 0)
	prMap := make(map[int]bool)
	for _, sha := range commits {
		ps, resp, err := global.Client.PullRequests.ListPullRequestsWithCommit(
			context.TODO(),
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			sha,
			nil,
		)
		if err != nil {
			global.Sugar.Errorw("load pr list",
				"call api", "failed",
				"err", err.Error(),
			)
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			global.Sugar.Errorw("load pr list",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			return nil
		}
		for _, v := range ps {
			// 同一个 pr 不重复记录
			if prMap[*v.Number] {
				continue
			}
			prs = append(prs, *v.Number)
			prMap[*v.Number] = true
		}
	}
	return prs
}

// 获取范围内所有 commit
func getRangeCommits(preSHA string) []string {
	// 只将第一行内容视为 SHA
	preSHA = strings.Split(strings.ReplaceAll(preSHA, "\r\n", "\n"), "\n")[0]
	commits := make([]github.RepositoryCommit, 0)
	page := 1
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	}

	for {
		tmp, resp, err := global.Client.Repositories.ListCommits(context.TODO(),
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			opt)
		if err != nil {
			global.Sugar.Errorw("load commit list",
				"call api", "failed",
				"err", err.Error(),
			)
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			global.Sugar.Errorw("load commit list",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			return nil
		}
		for _, v := range tmp {
			commits = append(commits, *v)
			// 已找到上次 commit
			if v.Parents[0].GetSHA() == preSHA {
				// 逆序 slice
				tmp := make([]string, len(commits))
				index := len(commits) - 1
				for _, v := range commits {
					tmp[index] = *v.SHA
					index--
				}
				return tmp
			}
			if len(commits) > 1000 {
				global.Sugar.Error("get commit list",
					"abnormal list length", len(commits))
				return nil
			}
		}
	}
}

// 获取最新的
func getLatestCommit() string {
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 1,
		},
	}
	commits, resp, err := global.Client.Repositories.ListCommits(
		context.TODO(),
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		opt,
	)
	if err != nil {
		global.Sugar.Errorw("load latest commit",
			"call api", "failed",
			"err", err.Error(),
		)
		return ""
	}
	if resp.StatusCode != http.StatusOK {
		global.Sugar.Errorw("load latest commit",
			"call api", "unexpect status code",
			"status", resp.Status,
			"status code", resp.StatusCode,
			"response", resp.Body,
		)
		return ""
	}
	if len(commits) == 0 {
		global.Sugar.Errorw("load latest commit",
			"call api", "unexpect resp length",
			"length", 0,
			"response", resp.Body,
		)
		return ""
	}
	return commits[0].GetSHA()
}

func getCommitIssue() *github.Issue {
	is, resp, err := global.Client.Issues.Get(context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		global.Conf.Repository.Spec.Workspace.CommitIssue,
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

func updateCommitIssue(is *github.Issue) {
	if is.Body == nil {
		global.Sugar.Errorw("update commit issue",
			"confirm", "failed",
			"cause", "body can not be nil",
			"data", is,
		)
		return
	}
	ir := issueToRequest(is)
	is, resp, err := global.Client.Issues.Edit(context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		*is.Number,
		ir,
	)
	if err != nil {
		global.Sugar.Errorw("update commit issue",
			"call api", "failed",
			"err", err.Error(),
		)
		return
	}
	if resp.StatusCode != http.StatusOK {
		global.Sugar.Errorw("update commit issue",
			"call api", "unexpect status code",
			"status", resp.Status,
			"status code", resp.StatusCode,
			"response", resp.Body,
		)
		return
	}
	global.Sugar.Info("update commit issue", "commit", is.Body)
}
