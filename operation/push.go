package operation

import (
	"fmt"
	client2 "issue-man/client"
)

// /pushed
func Push(i client2.Issue, c client2.Commenter) {
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

