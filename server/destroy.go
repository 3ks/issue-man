// destroy.go 对应 destroy 子命令的实现
// destroy 实现的是根据规则删除任务仓库的 issue。
package server

import (
	"context"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"issue-man/tools"
	"net/http"
	"sync"
	"time"
)

// Destroy 根据规则删除任务仓库的 issue。
func Destroy(conf config.Config) {
	issues, err := getIssues()
	if err != nil {
		global.Sugar.Errorw("Get issues files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}
	wg := sync.WaitGroup{}
	lock := make(chan int, 5)
	go func() {
		// 将 API 频率限制为每秒 5 次
		for range lock {
			time.Sleep(time.Millisecond * 200)
		}
	}()
	for _, v := range issues {
		wg.Add(1)
		go func(issue *github.Issue) {
			defer wg.Done()
			lock <- 1
			issueRequest := tools.Convert.Issue(issue)
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
