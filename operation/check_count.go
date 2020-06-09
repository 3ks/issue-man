package operation

import (
	"context"
	gg "github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"net/http"
)

// 数量检查
// 行为：根据配置的数量，检测评论人目前的对应状态的 issue 数量。
// 返回值为 true，则表示通过检测。
// 反之则表示未通过检测
func CheckCount(info Info, action *config.Action) bool {
	// nil 或 0 表示无限制
	if action.AddLabelsLimit == nil || *action.AddLabelsLimit <= 0 {
		return true
	}
	// 根据用户及 label 筛选
	req := &gg.IssueListByRepoOptions{}
	req.Assignee = info.Login

	for _, v := range action.AddLabels {
		req.Labels = append(req.Labels, *v)
	}

	is, resp, err := global.Client.Issues.ListByRepo(context.TODO(), info.Owner, info.Repository, req)
	if err != nil {
		global.Sugar.Errorw("CheckCount",
			"req_id", info.ReqID,
			"step", "call api",
			"info", info,
			"action", action,
			"req", req,
			"err", err.Error())
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("CheckCount",
			"req_id", info.ReqID,
			"step", "parse response",
			"info", info,
			"action", action,
			"req", req,
			"statusCode", resp.StatusCode,
			"body", string(body))
		return false
	}

	// 不超过 limit 限制
	if len(is) < *action.AddLabelsLimit {
		return true
	}
	return false
}
