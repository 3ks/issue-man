package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"net/http"
)

// 数量检查
// 行为：根据配置的数量，检测评论人目前的对应状态的 issue 数量。
// 返回值为 true，则表示通过检测。
// 反之则表示未通过检测
func CheckCount(info Info, labels []string, limit int) bool {
	// 0 表示无限制
	if limit <= 0 {
		return true
	}
	// 根据用户及 label 筛选
	req := &gg.IssueListByRepoOptions{}
	req.Assignee = info.Login
	req.Labels = labels

	is, resp, err := client.Get().Issues.ListByRepo(context.TODO(), info.Owner, info.Repository, req)
	if err != nil {
		fmt.Printf("list issue by repo fail. err: %v\n", err.Error())
		return false
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
		return false
	}

	// 不超过 limit 限制
	if len(is) < limit {
		return true
	}
	return false
}
