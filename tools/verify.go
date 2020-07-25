package tools

import (
	"context"
	gg "github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/global"
	"net/http"
)

// HasLabel
// 要求 require 的每一个元素都能在 source 中找到
func (v verifyFunctions) HasLabel(require, source []string) bool {
	if len(require) == 0 {
		return true
	}
	if len(source) == 0 {
		return false
	}

	sourceLabels := make(map[string]bool)
	for _, value := range source {
		sourceLabels[value] = true
	}

	for _, label := range require {
		if !sourceLabels[label] {
			return false
		}
	}
	return true
}

// HasAnyLabel
// 要求 require 的任意一个元素在 source 之中就可以
func (v verifyFunctions) HasAnyLabel(source []string, require ...string) bool {
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
		if requireLabels[value] {
			return true
		}
	}
	return false
}

const (
	Anyone     = "@anyone"
	Assignees  = "@assigner"
	Maintainer = "@maintainer"
	Member     = "@member"
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
	// 注意：默认只会 GET 100 个 issue，大多数情况下，都是够用的
	req := &gg.IssueListByRepoOptions{}
	req.Assignee = login
	req.Labels = labels
	req.PerPage = 100

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
