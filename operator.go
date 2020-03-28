package server

import (
	"fmt"
)

type Operator func(issue Issue, comment Commenter)

var ops map[string]Operator

// 单人可领取 Issue 数量限制
var MaxIssue int

// maintainer
var maintainer map[string]bool

// 指令
const (
	accept   = "/accept"
	push     = "/pushed"
	merged   = "/merged"
	assign   = "/assign"
	unassign = "/unassign"
)

// 状态
const (
	pending   = "status/pending"
	waiting   = "status/waiting-for-pr"
	reviewing = "status/reviewing"
	finish    = "status/merged"
)

func initOperator(c int) {
	MaxIssue = c

	ops = make(map[string]Operator)
	ops[accept] = Accept
	ops[push] = Push
	ops[merged] = Merge

	maintainer["a"] = true

	// 同类指令
	//
}

// /accept
func Accept(i Issue, c Commenter) {
	fmt.Printf("do: /accept, issue: %#v\n", i)

	comment := Comment{urls: i.urls}

	// 仅状态为 status/pending 的 issue 可以领取
	nowStatus := getStatus(i.Labels)
	if nowStatus != pending {
		// 忽略？提示？
		return
	}

	// 每个人可同时 /accept 的 issue 不能超过 MaxIssue 个
	// 不再执行分配操作分配
	if MaxIssue > 0 && TooManyIssue(i.urls.RepositoryURL+"/issues", c.Login) {
		comment.Body = fmt.Sprintf("Sorry @%s, We found that you have claimed %d issues, in order to ensure the quality of each issue, please complete those before claiming new one.", c.Login, MaxIssue)
		CommentIssue(comment)
		return
	}

	// 修改 label
	i.Labels = replaceStatus(i.Labels, pending, waiting)
	// 添加 assignees
	i.Assignees = append(i.Assignees, c.Login)
	if UpdateIssue(i) != nil {
		return
	}

	// 先修改，再提示
	comment.Body = fmt.Sprintf("Thank you @%s, this issue had been assigned to you.", c.Login)
	CommentIssue(comment)
}

// 限制 /accept 的 issue 数量
func TooManyIssue(url, login string) bool {
	return GetIssueCount(url, login) >= MaxIssue
}

// /pushed
func Push(i Issue, c Commenter) {
	// 仅状态为 status/translating 且评论人在 assignees 列表中才可生效
	nowStatus := getStatus(i.Labels)
	if nowStatus != waiting || loginNotInAssignees(i.Assignees, c.Login) {
		// 忽略？提示？
		return
	}
	fmt.Printf("do: /pushed, issue: %#v\n", i)

	// 修改 label
	i.Labels = replaceStatus(i.Labels, waiting, reviewing)
	UpdateIssue(i)
}

// 分配和取消分配，可在一条请求内完成
// 策略：
// 普通成员不能进行 assign 操作，但可以通过 /accept 将 pending 状态的 issue 分配给自己，或者由 maintainer 分配。
// 普通成员只能 unassign 已经分配给自己的 issue
// maintainer 可以进行 assign 和 unassign 分配操作。
func Assign(i Issue, c Commenter) {
	// 仅状态为 status/translating 且评论人在 assignees 列表中才可生效
	nowStatus := getStatus(i.Labels)
	if nowStatus != waiting || loginNotInAssignees(i.Assignees, c.Login) {
		// 忽略？提示？
		return
	}
	fmt.Printf("do: /pushed, issue: %#v\n", i)

	// 修改 label
	i.Labels = replaceStatus(i.Labels, waiting, reviewing)
	UpdateIssue(i)
}

// /merged
func Merge(i Issue, c Commenter) {
	// 仅状态为 status/reviewing 且评论人在 assignees 列表中才可生效
	nowStatus := getStatus(i.Labels)
	if nowStatus != reviewing || loginNotInAssignees(i.Assignees, c.Login) {
		return
		// 忽略？提示？
	}

	fmt.Printf("do: /merged, issue: %#v\n", i)

	// 修改 label
	i.Labels = replaceStatus(i.Labels, reviewing, finish)
	i.State = "closed"
	UpdateIssue(i)
}

func loginNotInAssignees(assignees []string, login string) bool {
	for _, v := range assignees {
		if v == login {
			return false
		}
	}
	return true
}

func getStatus(s []string) string {
	for _, v := range s {
		switch v {
		case pending:
			return pending
		case waiting:
			return waiting
		case reviewing:
			return reviewing
		case finish:
			return finish
		}
	}
	return ""
}

func replaceStatus(s []string, old, new string) []string {
	for k, v := range s {
		if v == old {
			s[k] = new
			return s
		}
	}
	return s
}
