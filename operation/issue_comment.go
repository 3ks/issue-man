package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"net/http"
)

// 尝试 comment issue
func IssueComment(info Info, body string) {
	// 如果 body 为空，则不做任何操作
	if body == "" {
		return
	}

	comment := &gg.IssueComment{}
	comment.Body = &body
	_, resp, err := client.Get().Issues.CreateComment(context.TODO(), info.Owner, info.Repository, info.IssueNumber, comment)
	if err != nil {
		fmt.Printf("comment_issue_fail err: %v\n", err.Error())
		return
	}

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("comment_issue_maybe_fail status_code: %v\n", resp.StatusCode)
		return
	}
}
