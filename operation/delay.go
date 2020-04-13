package operation

import (
	"fmt"
	"time"
)

// 获取最终重置日期
// 根据 event 获取原定重置日期
// 根据指令的日期和延迟的时长，计算期望重置日期
// 返回较大的那个日期
func getResetDate(owner, repository string, issueNumber int, labels []string, in int64, delay int64) string {
	createdAt, err := getLabelCreateAt(owner, repository, issueNumber, labels)
	if err != nil {
		fmt.Printf("get label create at failed. label: %v, err: %v\n", labels, err.Error())
		return ""
	}
	// 粗略算作执行指令的时间
	exceptedTime := createdAt.AddDate(0, 0, int(in))
	delayTime := time.Now().AddDate(0, 0, int(delay))
	if exceptedTime.Sub(delayTime) > 0 {
		return exceptedTime.Format("2006-01-02")
	}
	return delayTime.Format("2006-01-02")
}
