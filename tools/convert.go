package tools

import (
	"github.com/google/go-github/v30/github"
	"sort"
)

// MapToString
// 将 map 转换为 slice
func (c convertFunctions) Issue(issue *github.Issue) (ir *github.IssueRequest) {
	ir = &github.IssueRequest{
		Title: issue.Title,
		Body:  issue.Body,
		State: issue.State,
		//Milestone: issue.Milestone.Number,
		//Labels:    convertLabel(issue.Labels),
		//Assignees: convertAssignees(issue.Assignees),
	}
	if issue.Milestone != nil {
		ir.Milestone = issue.Milestone.Number
	}
	if issue.Labels != nil {
		ir.Labels = c.Label(issue.Labels)
	}
	if issue.Assignees != nil {
		ir.Assignees = c.Assignees(issue.Assignees)
	}
	return
}

// Label
// 传入 github.Issue，返回 *[]string
// 一般用于将 github.Issue.Label 转换为 github.IssueRequest.Label
func (c convertFunctions) Label(sourceLabel []*github.Label) *[]string {
	if sourceLabel == nil {
		return nil
	}
	labels := make([]string, len(sourceLabel))
	count := 0
	for _, v := range sourceLabel {
		labels[count] = v.GetName()
		count++
	}
	return &labels
}

// MapToString
// 将 map 转换为 slice
func (c convertFunctions) Assignees(sourceUser []*github.User) *[]string {
	if sourceUser == nil {
		return nil
	}
	assignees := make([]string, len(sourceUser))
	for k, v := range sourceUser {
		assignees[k] = v.GetLogin()
	}
	return &assignees
}

// LabelAdd
// 由于 GitHub 对于多个重名 Label 可能会重复创建，
// 所以应该用该函数对 Label 进行去重添加
func (c convertFunctions) LabelAdd(label *[]string, add ...string) *[]string {
	if label == nil {
		newSlice := make([]string, 0)
		label = &newSlice
	}
	tmp := make(map[string]bool)
	for _, v := range *label {
		tmp[v] = true
	}
	for _, v := range add {
		tmp[v] = true
	}

	return c.MapToString(tmp)
}

// LabelRemove
// 用于移除指定的 label
func (c convertFunctions) LabelRemove(label *[]string, remove ...string) *[]string {
	if label == nil {
		return nil
	}
	tmp := make(map[string]bool)
	for _, v := range *label {
		tmp[v] = true
	}
	for _, v := range remove {
		delete(tmp, v)
	}

	return c.MapToString(tmp)
}

// MapToString
// 将 map 转换为 slice
func (c convertFunctions) MapToString(source map[string]bool) (array *[]string) {
	tmp := make([]string, len(source))
	index := 0
	for k := range source {
		tmp[index] = k
		index++
	}
	sort.Strings(tmp)
	return &tmp
}
