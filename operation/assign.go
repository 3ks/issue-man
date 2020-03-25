package operation

import (
	"fmt"
	client2 "issue-man/client"
)

// 分配和取消分配，可在一条请求内完成
// 策略：
// 普通成员不能进行 assign 操作，但可以通过 /iAccept 将 sPending 状态的 issue 分配给自己，或者由 maintainer 分配。
// 普通成员只能 unassign 已经分配给自己的 issue
// maintainer 可以进行 assign 和 unassign 分配操作。
func Assign(i client2.Issue, c client2.Commenter) {
	// 仅状态为 status/translating 且评论人在 assignees 列表中才可生效
	nowStatus := getStatus(i.Labels)
	if nowStatus != SWaiting || loginNotInAssignees(i.Assignees, c.Login) {
		// 忽略？提示？
		return
	}
	fmt.Printf("do: /pushed, issue: %#v\n", i)

	// 修改 label
	i.Labels = replaceStatus(i.Labels, SWaiting, SReviewing)
	client2.UpdateIssue(i)
}
