package operation

import (
	"fmt"
	"issue-man/server"
)

// /pushed
func Push(i server.Issue, c server.Commenter) {
	// 仅状态为 status/translating 且评论人在 assignees 列表中才可生效
	nowStatus := getStatus(i.Labels)
	if nowStatus != SWaiting || loginNotInAssignees(i.Assignees, c.Login) {
		// 忽略？提示？
		return
	}
	fmt.Printf("do: /pushed, issue: %#v\n", i)

	// 修改 label
	i.Labels = replaceStatus(i.Labels, SWaiting, SReviewing)
	server.UpdateIssue(i)
}

