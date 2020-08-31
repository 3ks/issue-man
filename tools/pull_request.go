package tools

import (
	"context"
	"github.com/google/go-github/v30/github"
	"issue-man/global"
	"net/http"
)

// 获取最近一次 merged 的 pull request
func (i pullRequestFunctions) LatestMerged() (latestPR *github.PullRequest) {
	opt := &github.PullRequestListOptions{
		State: "close",
		Base:  global.Conf.Repository.Spec.Source.Branch,
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
			if v.MergeCommitSHA != nil {
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
func (i pullRequestFunctions) ListRangePRs(prNumber int) (prs []*github.PullRequest) {
	if prNumber == 0 {
		return nil
	}
	prs = make([]*github.PullRequest, 0)
	// 逆序，先处理较早的 pr，最后处理最新的 pr
	// 最后一个元素是最近一个 pr
	defer func() {
		Convert.ReversePR(prs)
		global.Sugar.Infow("get valid pull requests", "len", len(prs))
	}()

	opt := &github.PullRequestListOptions{
		State: "closed",
		Base:  "master",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
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
			// 已遍历完需要检测的 pr
			if v.GetNumber() <= prNumber {
				return prs
			}

			// 仅处理 merged 的 pr
			if v.MergeCommitSHA != nil {
				prs = append(prs, v)
				global.Sugar.Debugw("get valid pull request", "number", v.GetNumber())
			}
			// 避免极端情况下，需要在调用一次 API 的情况
			if v.GetNumber()-1 <= prNumber {
				return prs
			}
		}
		// 不太可能出现
		if len(ps) < 100 {
			break
		}
		opt.Page++
	}
	return prs
}
