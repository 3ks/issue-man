package server

import (
	"context"
	"expvar"
	"github.com/google/go-github/v30/github"
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
		time.Now().Format("2006-01-02 ")+*global.Conf.IssueCreate.Spec.DetectionAt,
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
	preCommit := getCommitIssue()
	if preCommit == nil {
		return
	}
	// 获取 commit 列表
	commits := getRangeCommits(*preCommit.Body)
	preCommit.Body = &commits[len(commits)-1]
	// 更新 commit issue
	defer updateCommitIssue(preCommit)

	// 获取 pr 列表
	prs := getAssociatedPRs(commits)

	// 获取 pr 涉及的文件列表
	files := getAssociatedFiles(prs)
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
	for _, v := range files {
		wg.Add(1)
		go func(file File) {
			defer wg.Done()
			// 1. 判断是否需要处理
			// 2. 判断是否有已存在的 issue
			//    存在则更新
			//       对于删除操作，删除后文件为 0 的加 label 处理
			//       对于移动操作，需要两个操作，删除源 issue，添加（或更新）至目的 issue
			//    不存在则新增
			// TODO 精确 diff

			if err != nil {
				global.Sugar.Errorw("get issues files",
					"status", "fail",
					"err", err.Error(),
				)
				return
			}
		}(v)
	}
	wg.Wait()
}

type File struct {
	PrNumber int `json:"pr_number"`
	cf       *github.CommitFile
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
				global.Conf.Repository.Spec.Upstream.Owner,
				global.Conf.Repository.Spec.Upstream.Repository,
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
			global.Conf.Repository.Spec.Upstream.Owner,
			global.Conf.Repository.Spec.Upstream.Repository,
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
		Path: "",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 100,
		},
	}

	for {
		tmp, resp, err := global.Client.Repositories.ListCommits(context.TODO(),
			global.Conf.Repository.Spec.Upstream.Owner,
			global.Conf.Repository.Spec.Upstream.Repository,
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

func getCommitIssue() *github.Issue {
	is, resp, err := global.Client.Issues.Get(context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		*global.Conf.Repository.Spec.CommitIssue,
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
	global.Sugar.Errorw("update commit issue", "commit", is.Body)
}
