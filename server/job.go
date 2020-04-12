package server

import (
	"fmt"
	"issue-man/config"
	"issue-man/operation"
	"time"
)

var Jobs []config.Job

// job
// 目前主要完成状态持续时间的检测，并提醒
// 思路：对于需要检测的状态（label），会将其添加至相应的切片
//      每天定时检测，满足相关条件时，则执行一些操作
func job(conf config.Config) {
	// 获取需要检测的状态
	Jobs = make([]config.Job, 0)
	for _, v := range config.Instructions {
		for _, job := range v.Jobs {
			// 小于 0 的会被忽略掉
			if job.In < 0 {
				continue
			}
			Jobs = append(Jobs, job)
		}
	}

	// 无任务
	if len(Jobs) == 0 {
		return
	}
	fmt.Printf("loaded jobs: %#v\n", Jobs)

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
			for _, v := range Jobs {
				operation.Job(conf.FullRepositoryName, v)
			}
		}
		// 等待明天的这个时刻
		s = t.AddDate(0, 0, 1).Sub(time.Now())
		fmt.Printf("waiting for to detection. time: %v\n", s.String())
		time.Sleep(s)
	}
}
