package operation

import (
	"issue-man/comm"
	"issue-man/global"
	"issue-man/tools"
	"sync"
	"time"
)

var (
	// SyncIssues() 可以通过多种方式触发
	// 这里加一个锁，以避免重复检测提示的情况
	lock sync.Mutex
)

// Sync
// 目前主要完成状态持续时间的检测，并提醒
// 思路：对于需要检测的状态（label），会将其添加至相应的切片
//      每天定时检测，满足相关条件时，则执行一些操作
//
// Sync 用于定时自动调用同步检测函数
// TODO 检测频率
// 1. 获取所有特定 label 的 issue
// 2. 获取存储 commit 的 issue
// 3. 遍历 commit，存储到栈内，直至第二步匹配的 commit。
// 4. pop commit 栈，分析涉及的文件，是否存在匹配的 issue
// 5. 对匹配的 issue，comment 提示，该 issue 对应的某个文件在哪次 commit 有变动
func Sync() {
	global.Sugar.Info("loaded jobs", "list", global.Jobs)
	// 解析检测时间
	t, err := time.ParseInLocation("2006-01-02 15:04",
		time.Now().Format("2006-01-02 ")+global.Conf.Repository.Spec.Workspace.Detection.At,
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
		// 同步检测是一个特殊的任务，会检测两次 pr 之间所有 merged pr 涉及的文件，并提示
		SyncIssues()

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

// SyncIssues 同步检测 issue
func SyncIssues() {
	// SyncIssues 可以通过多种方式触发
	// 这里加一个锁，以避免重复检测提示的情况
	lock.Lock()
	defer lock.Unlock()

	// 获取 pr issue
	prIssue := tools.Issue.GetPRIssue()
	if prIssue == nil {
		return
	}

	// 获取 pr 列表
	prs, headSha := tools.PR.ListRangePRs(tools.Parse.PRNumberFromBody(prIssue.GetBody()))
	// 更新 pr issue
	defer tools.Issue.Edit(prIssue)
	// 最近一次 pr
	// 如果中途失败，需要再次生成 body，以保存进度
	prIssue.Body = tools.Generate.BodyByPRNumberAndSha(prs[len(prs)-1], headSha)

	// 获取每个 pr 涉及的文件列表
	files := tools.PR.GetAssociatedFiles(prs)

	// 获取现有 issue 列表
	existIssues, err := tools.Issue.GetAllMath()
	if err != nil {
		global.Sugar.Errorw("Get issues files",
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
		go func(file comm.File) {
			defer wg.Done()
			// 1. 判断是否需要处理
			for _, include := range global.Conf.IssueCreate.Spec.Includes {
				// 符合条件的文件
				if global.Conf.IssueCreate.SupportFile(include, file.CommitFile.GetFilename()) {
					lock <- 1
					file.Sync(
						include,
						existIssues[*tools.Generate.Title(file.CommitFile.GetFilename())],
						existIssues[*tools.Generate.Title(file.CommitFile.GetPreviousFilename())],
					)
					// 文件已处理
					return
				}
			}
		}(file)
	}
	wg.Wait()
}
