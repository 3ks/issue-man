package server

import (
	"fmt"
	"issue-man/config"
	"issue-man/global"
	"issue-man/operation"
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
	fmt.Printf("loaded jobs: %#v\n", global.Jobs)
	// 无任务
	if len(global.Jobs) == 0 {
		return
	}

	// 解析检测时间
	t, err := time.ParseInLocation("2006-01-02 15:04", time.Now().Format("2006-01-02 ")+conf.DetectionAt, time.Local)
	if err != nil {
		fmt.Printf("can't pasrse detection time. start job failed.\n")
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
	fmt.Printf("waiting for to detection. time: %v\n", s.String())
	time.Sleep(s)

	for {
		skip := false
		// 周末放假
		if conf.SkipWeekend {
			switch time.Now().Weekday() {
			case time.Sunday:
				skip = true
			default:
				skip = false
			}
		}
		// 遍历检测任务
		if !skip {
			for _, v := range global.Jobs {
				operation.Job(conf.FullRepositoryName, v)
			}
		}
		// 等待明天的这个时刻
		t = t.AddDate(0, 0, 1)
		s = t.Sub(time.Now())
		fmt.Printf("waiting for to detection. time: %v\n", s.String())
		time.Sleep(s)
	}
}

// 521
// 21
// 1
