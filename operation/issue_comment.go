package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/global"
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
	_, resp, err := global.Client.Issues.CreateComment(context.TODO(), info.Owner, info.Repository, info.IssueNumber, comment)
	if err != nil {
		fmt.Printf("comment_issue_fail err: %v\n", err.Error())
		global.Sugar.Errorw("IssueComment",
			"req_id", info.ReqID,
			"step", "call api",
			"status", "fail",
			"info", info,
			"err", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("CheckCount",
			"req_id", info.ReqID,
			"step", "parse response",
			"info", info,
			"statusCode", resp.StatusCode,
			"body", string(body))
		return
	}
}
