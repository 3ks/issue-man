package operation

import (
	"fmt"
	"time"
)

// 获取最终重置日期
// 根据 event 获取 label 创建时间，计算出原定重置日期
// 根据最后执行延迟指令的日期和延迟的时长，计算期望重置日期
// 返回较大的那个日期
func getResetDate(owner, repository string, issueNumber int, labels []string, in, delay int, instructName string) (resetDate time.Time, err error) {
	createdAt, err := getLabelCreateAt(owner, repository, issueNumber, labels)
	if err != nil {
		fmt.Printf("get label create at failed. label: %v, err: %v\n", labels, err.Error())
		return resetDate, err
	}
	// 原定重置时间
	exceptedTime := createdAt.AddDate(0, 0, in)

	if delay == 0 && instructName == "" {
		return exceptedTime, nil
	}

	// 计算延迟后的重置时间
	delayDate, err := LastDelayAt(owner, repository, issueNumber, delay, instructName)
	if err != nil {
		fmt.Printf("can found last delay time. err: %v\n", err.Error())
		return exceptedTime, nil
	}
	// 返回较大的日期
	if exceptedTime.Sub(delayDate) > 0 {
		return exceptedTime, nil
	} else {
		return delayDate, nil
	}
}
