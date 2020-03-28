package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"issue-man/instruction"
	"net/http"
	"strings"
)

const (
	IssueOpen   = "open"
	IssueClosed = "closed"
)

// 对 issue 进行修改
// 具体修改内容完全取决于配置文件
// 但是，一般来说，改动的内容只涉及 label，assignees，state
// 而 title，body，milestone 不会改变
func IssueEdit(info Info, flow instruction.Flow) {
	// 一般不会变化的内容
	req := &gg.IssueRequest{
		Title:     &info.Title,
		Body:      &info.Body,
		Milestone: &info.Milestone,
	}

	closeIssue := IssueOpen
	// 是否关闭 issue
	if flow.Close {
		closeIssue = IssueClosed
	}
	req.State = &closeIssue

	// 更新 label（如果有的话）
	updateLabel(req, info, flow)

	// 更新 assignees（如果有的话）
	updateAssign(req, info, flow)

	// 尝试调用更新接口
	_, resp, err := client.Get().Issues.Edit(context.TODO(), info.Owner, info.Repository, info.IssueNumber, req)
	if err != nil {
		fmt.Printf("update_isse_fail err: %v\n", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("update_isse_maybe_fail status_code: %v\n", resp.StatusCode)
		return
	}

	// 创建文本提示
	if flow.SuccessFeedback != "" {
		commentBody := strings.ReplaceAll(flow.SuccessFeedback, "@somebody", info.Login)
		IssueComment(info, commentBody)
	}
}

// 根据 flow 更新 info 中的 label
func updateLabel(req *gg.IssueRequest, info Info, flow instruction.Flow) {
	current := make(map[string]bool)
	for _, v := range flow.CurrentLabel {
		current[v] = true
	}

	labels := make([]string, 0)
	labels = append(labels, flow.TargetLabel...)
	for _, v := range info.Labels {
		if _, ok := current[v]; ok {
			continue
		}
		labels = append(labels, v)
	}
	req.Labels = &labels
}

// 根据 flow 更新 info 中的 assignees
func updateAssign(req *gg.IssueRequest, info Info, flow instruction.Flow) {
	switch flow.Mention {
	case "addition":
		info.Assignees = append(info.Assignees, info.Mention...)
	case "remove":
		removeAssign(req, info)
	default:
		return
	}
}

func removeAssign(req *gg.IssueRequest, info Info) {
	newAssignees := make([]string, 0)
	rm := make(map[string]bool)
	for _, v := range info.Mention {
		rm[v] = true
	}
	for _, v := range info.Assignees {
		if _, ok := rm[v]; ok {
			continue
		}
		newAssignees = append(newAssignees, v)
	}
	req.Assignees = &newAssignees
}
