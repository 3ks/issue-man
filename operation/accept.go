package operation

import (
	"fmt"
	"issue-man/server"
)

// Accept
func Accept(i server.Issue, c server.Commenter) {
	fmt.Printf("do: /iAccept, issue: %#v\n", i)

	comment := server.Comment{URL: i.URL}

	// 仅状态为 status/sPending 的 issue 可以领取
	nowStatus := getStatus(i.Labels)
	if nowStatus != SPending {
		// 忽略？提示？
		return
	}

	// 每个人可同时 /iAccept 的 issue 不能超过 MaxIssue 个
	// 不再执行分配操作分配
	if MaxIssue > 0 && tooManyIssue(i.URL.RepositoryURL+"/issues", c.Login) {
		comment.Body = fmt.Sprintf("Sorry @%s, We found that you have claimed %d issues, in order to ensure the quality of each issue, please complete those before claiming new one.", c.Login, MaxIssue)
		server.CommentIssue(comment)
		return
	}

	// 修改 label
	i.Labels = replaceStatus(i.Labels, SPending, SWaiting)
	// 添加 assignees
	i.Assignees = append(i.Assignees, c.Login)
	if server.UpdateIssue(i) != nil {
		return
	}

	// 先修改，再提示
	comment.Body = fmt.Sprintf("Thank you @%s, this issue had been assigned to you.", c.Login)
	server.CommentIssue(comment)
}

