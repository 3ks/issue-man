// destroy.go 对应 destroy 子命令的实现
// destroy 实现的是根据规则删除任务仓库的 issue。
package server

import (
	"github.com/google/go-github/v30/github"
	"issue-man/config"
	"issue-man/global"
	"issue-man/tools"
	"sync"
	"time"
)

// Destroy 根据规则删除任务仓库的 issue。
func Destroy(conf config.Config) {
	issues, err := tools.Issue.GetAllMath()
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
		// 将 API 频率限制为每秒 2 次
		for range lock {
			time.Sleep(time.Millisecond * 100)
		}
	}()
	for _, v := range issues {
		wg.Add(1)
		go func(issue *github.Issue) {
			defer wg.Done()
			lock <- 1
			state := "close"
			issue.State = &state
			_, _ = tools.Issue.Edit(issue)
		}(v)
	}
	wg.Wait()

	global.Sugar.Infow("destroy issues",
		"status", "done")
}
