package tools

import (
	"context"
	"github.com/google/go-github/v30/github"
	"issue-man/comm"
	"issue-man/global"
	"net/http"
	"sync"
)

// 获取最近一次 merged 的 pull request
func (i pullRequestFunctions) LatestMerged() (latestPR *github.PullRequest) {
	opt := &github.PullRequestListOptions{
		State: "close",
		Base:  "master",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 30,
		},
	}
	for {

		prs, resp, err := global.Client.PullRequests.List(
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
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			global.Sugar.Errorw("load latest commit",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			return nil
		}
		for _, v := range prs {
			// 最近的一个 merged pr
			if v.GetMerged() {
				return v
			}
		}
		if len(prs) < opt.PerPage {
			break
		}
		opt.Page++
	}
	return nil
}

// 获取从给定 pr number 到最近一次 pr 之间所有的 merged pr
// 以及最近一次 merged pr 的 commit sha
func (i pullRequestFunctions) ListRangePRs(prNumber int) (prs []int, headSha string) {
	if prNumber == 0 {
		return nil, ""
	}
	prs = make([]int, 0)
	// 逆序，先处理较早的 pr，最后处理最新的 pr
	defer Convert.Reverse(prs)

	opt := &github.PullRequestListOptions{
		State: "closed",
		Base:  "master",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
	once := sync.Once{}
	//
	for {
		ps, resp, err := global.Client.PullRequests.List(
			context.TODO(),
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			opt,
		)
		if err != nil {
			global.Sugar.Errorw("load pr list",
				"call api", "failed",
				"err", err.Error(),
			)
			return nil, ""
		}
		if resp.StatusCode != http.StatusOK {
			global.Sugar.Errorw("load pr list",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			return nil, ""
		}
		for _, v := range ps {
			// 已遍历完需要检测的 pr
			if v.GetNumber() <= prNumber {
				return prs, headSha
			}
			// 仅处理 merged 的 pr
			if v.GetMerged() {
				prs = append(prs, v.GetNumber())
				// 记录最近一次 merged pr 的 commit sha
				once.Do(func() {
					headSha = v.GetMergeCommitSHA()
				})
			}
		}
		// 不太可能出现
		if len(ps) < 100 {
			break
		}
		opt.Page++
	}
	return prs, headSha
}

func (i pullRequestFunctions) GetAssociatedFiles(prs []int) []comm.File {
	files := make([]comm.File, 0)

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
				files = append(files, comm.File{
					PrNumber:   v,
					CommitFile: cf,
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
