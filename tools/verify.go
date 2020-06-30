package tools

import (
	"context"
	gg "github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/global"
	"net/http"
)

// HasLabel
// 要求 require 的每一个元素都在 source 之中
func (v verifyFunctions) HasLabel(source []string, require ...string) bool {
	if len(require) == 0 {
		return true
	}
	if len(source) == 0 {
		return false
	}

	requireLabels := make(map[string]bool)
	for _, value := range require {
		requireLabels[value] = true
	}

	for _, value := range source {
		if !requireLabels[value] {
			return false
		}
	}
	return true
}

const (
	Anyone     = "anyone"
	Assignees  = "assignees"
	Maintainer = "maintainers"
	Member     = "member"
)

// 权限检查
// 行为：根据检查策略，对评论人进行权限检查
// 返回值为 true，则表示通过检测。
// 反之则表示未通过检测。
// 检查流程是，先检测配置文件是否配置了改项，如果配置了，用户是否满足该项的条件。
// 满足任一一个条件，则视为有权限。
// permission 可填的值目前有：anyone、assignees、maintainers、member
func (v verifyFunctions) Permission(permission []string, login string, assignees []string) bool {
	// 未配置任何权限，则不允许操作
	if len(permission) == 0 {
		return false
	}

	// 要求的权限 map
	permissionMap := Convert.StringToMap(permission)
	// 当前的 assignees map
	assigneesMap := Convert.StringToMap(assignees)

	// maintainer 可以操作
	if permissionMap[Maintainer] && global.Maintainers[login] {
		return true
	}

	// assigner 可以操作
	// 自身在当前 assignees 列表中
	if permissionMap[Assignees] && assigneesMap[login] {
		return true
	}

	// member 可以操作
	if permissionMap[Member] && global.Members[login] {
		return true
	}

	// anyone 可以操作
	if permissionMap[Anyone] {
		return true
	}

	return false
}

// 数量检查
// 行为：根据配置的数量，检测评论人目前的对应状态的 issue 数量。
// 返回值为 true，则表示通过检测。
// 反之则表示未通过检测
func (v verifyFunctions) LabelCount(login string, labels []string, limit int) bool {
	// 小于等于 0 为无限制
	if limit <= 0 {
		return true
	}

	// 根据用户及 label 筛选
	// TODO 默认只 GET 30 个 issue
	req := &gg.IssueListByRepoOptions{}
	req.Assignee = login
	req.Labels = labels

	is, resp, err := global.Client.Issues.ListByRepo(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		req)
	if err != nil {
		global.Sugar.Errorw("CheckCount",
			"step", "call api",
			"login", login,
			"labels", labels,
			"limit", limit,
			"req", req,
			"err", err.Error())
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("CheckCount",
			"step", "parse response",
			"login", login,
			"labels", labels,
			"limit", limit,
			"req", req,
			"statusCode", resp.StatusCode,
			"body", string(body))
		return false
	}

	// 超过 limit 限制
	if len(is) > limit {
		return false
	}

	return true
}
