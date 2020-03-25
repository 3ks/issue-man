package operator

import (
	"fmt"
	"issue-man/server"
)

// Merge
func Merge(i server.Issue, c server.Commenter) {
	// 仅状态为 status/reviewing 且评论人在 assignees 列表中才可生效
	nowStatus := getStatus(i.Labels)
	if nowStatus != SReviewing || loginNotInAssignees(i.Assignees, c.Login) {
		return
		// 忽略？提示？
	}

	fmt.Printf("do: /iMerged, issue: %#v\n", i)

	// 修改 label
	i.Labels = replaceStatus(i.Labels, SReviewing, SFinish)
	i.State = "closed"
	server.UpdateIssue(i)
}

