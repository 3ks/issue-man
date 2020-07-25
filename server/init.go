// init.go 对应 init 子命令的实现
// init 实现的是根据上游仓库和规则，在任务仓库创建初始化 issue
package server

import (
	"github.com/google/go-github/v30/github"
	"issue-man/config"
	"issue-man/global"
	"issue-man/tools"
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
	lock <- 1
	defer func() {
		<-lock
	}()
	// init 始终基于最新 pr 来完成，
	// 所以这里直接更新 pr issue body
	//prIssue := tools.Issue.GetPRIssue()
	latestPR := tools.PR.LatestMerged()
	//prIssue.Body = tools.Generate.BodyByPRNumberAndSha(latestPR.GetNumber(), latestPR.GetMergeCommitSHA())
	//defer tools.Issue.Edit(prIssue)
	genAndCreateIssues(latestPR.GetMergeCommitSHA())
}

// 根据配置、文件列表、已存在 issue，判断生成最终操作列表
// 遍历文件，判断文件是否符合条件，符合则直接创建
// （实现根据文件名生成 title、body。根据配置生成 label、assignees 的 issue）
// （实现根据 body 提取文件列表的方法）
// 1. 根据规则获取已存在 issue 列表
// 2. 遍历，根据 title 判断 issue 是否已经存在
// 3. 更新 issue（如果文件有变化），assignees 如果不为空，则不修改，如果为空则判断配置是否有配置 assignees，如都为空则不操作。
// 4. 创建 issue
func genAndCreateIssues(sha string) {
	// 获取全部需要处理的文件
	fs, err := tools.Tree.GetAllMatchFile(sha)
	if err != nil {
		global.Sugar.Errorw("Get upstream files",
			"status", "fail",
			"err", err.Error(),
		)
		return
	}

	// 获取全部符合条件的 issue，避免重复创建
	existIssues, err := tools.Issue.GetAllMath()
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

	// 根据配置和已有 issue 判断是创建或更新
	for file := range fs {
		include, ok := global.Conf.IssueCreate.SupportFile(file)
		// 符合条件的文件
		if ok {
			// 根据 title 判断，如果已存在相关 issue，则更新
			exist := existIssues[*tools.Generate.Title(file, include)]
			if exist != nil {
				updates[*exist.Number] = tools.Generate.UpdateIssue(false, file, *exist)
			} else {
				// 不存在，则新建，新建也分两种情况
				// 有多个新文件属于一个 issue
				create := creates[*tools.Generate.Title(file, include)]
				if create != nil {
					creates[*tools.Generate.Title(file, include)] = tools.Generate.UpdateNewIssue(file, create)
				} else {
					// 是一个新的新 issue
					creates[*tools.Generate.Title(file, include)] = tools.Generate.NewIssue(include, file)
				}
			}
		}
	}
	global.Sugar.Debugw("create issues", "data", creates)
	global.Sugar.Debugw("update issues", "data", updates)

	wg := sync.WaitGroup{}
	lock := make(chan int, 5)
	go func() {
		// 由于 GitHub 的额外限制
		// 将 API 频率限制为每秒 2 次
		for range lock {
			time.Sleep(time.Millisecond * 500)
		}
	}()
	// update 的 issue
	for k, v := range updates {
		wg.Add(1)
		go func(number int, issue *github.IssueRequest) {
			defer wg.Done()
			lock <- 1
			_, err := tools.Issue.EditByIssueRequest(number, issue)
			if err != nil {
				updateFail++
				return
			}
		}(k, v)
	}
	wg.Wait()

	// create 的 issue
	for _, v := range creates {
		wg.Add(1)
		go func(issue *github.IssueRequest) {
			defer wg.Done()
			lock <- 1
			_, err := tools.Issue.Create(issue)
			if err != nil {
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
