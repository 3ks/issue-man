package comm

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Commenter = "@commenter"
	Count     = "@count"
	ResetDate = "@reset-date"
	ReqID     = "@req-id"
	Assignees = "@assignees"
)

// 替换文本提示里的特殊字符
// 后面可能会再增加一些特殊字符
type Comment struct {
	Login      string
	LimitCount int
	ResetDate  string
	ReqID      string
	Assignees  []string
}

// 这里只对一些关键字做替换
// 具体的值需要自行计算
func (r Comment) HandComment(text string) string {
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, Commenter, fmt.Sprintf("@%s", r.Login))
	text = strings.ReplaceAll(text, Count, strconv.Itoa(r.LimitCount))
	text = strings.ReplaceAll(text, ResetDate, fmt.Sprintf("`%s`", r.ResetDate))
	text = strings.ReplaceAll(text, ReqID, fmt.Sprintf("`%s`", r.ReqID))
	if strings.Contains(text, Assignees) {
		all := ""
		for k, v := range r.Assignees {
			if k+1 == len(r.Assignees) {
				all = fmt.Sprintf("%s@%s ", all, v)
			} else {
				all = fmt.Sprintf("%s@%s, ", all, v)
			}
		}
		text = strings.ReplaceAll(text, Assignees, all)
	}
	return text
}
