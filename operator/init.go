package operator

import (
	"issue-man/model"
	"issue-man/server"
	"strings"
)

// 每个操作的格式
type Operator func(issue server.Issue, comment server.Commenter)

// 支持的操作列表
var ops map[string]Operator

// 单人可领取 Issue 数量限制
var MaxIssue int

// maintainer
var maintainer map[string]bool

// 指令
const (
	IAccept  = "/accept"
	IPush    = "/pushed"
	IMerged  = "/merged"
	IAssign   = "/assign"
	IUnassign = "/unassign"
)

// 状态
const (
	SPending  = "status/spending"
	SWaiting   = "status/waiting-for-pr"
	SReviewing = "status/reviewing"
	SFinish    = "status/merged"
)

func Handing(i model.IssueHook,c model.IssueQuery){
	// split \n ?
	// 有对应指令则执行对应操作
	// todo 先提取指令，再调用方法
	// todo 对于同一类指令，这样只需要调用一次
	// todo 例如：多个 /cc，/unassign 和 /assign 同时调用等。
	for k, v := range ops {
		if strings.Contains(c.Body, k) {
			v(i, c)
			break
		}
	}
}
func InitOperator(c int) {
	MaxIssue = c

	ops = make(map[string]Operator)
	ops[IAccept] = Accept
	ops[IPush] = Push
	ops[IMerged] = Merge
	ops[IAssign] = Assign
	ops[IUnassign] = Assign

	// 指定 maintainer
	maintainer["a"] = true
}

// 限制 Accept 的 issue 数量
func tooManyIssue(url, login string) bool {
	return server.GetIssueCount(url, login) >= MaxIssue
}

// 判断评论中是 assigner
// todo 或者 master
func loginNotInAssignees(assignees []string, login string) bool {
	for _, v := range assignees {
		if v == login {
			return false
		}
	}
	return true
}

// 判断 issue 状态
func getStatus(s []string) string {
	for _, v := range s {
		switch v {
		case SPending:
			return SPending
		case SWaiting:
			return SWaiting
		case SReviewing:
			return SReviewing
		case SFinish:
			return SFinish
		}
	}
	return ""
}

// 修改状态
func replaceStatus(s []string, old, new string) []string {
	for k, v := range s {
		if v == old {
			s[k] = new
			return s
		}
	}
	return s
}
