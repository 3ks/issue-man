package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"issue-man/config"
	"net/http"
	"time"
)

const (
	IssueOpen   = "open"
	IssueClosed = "closed"
)

// 对 issue 进行修改
// 具体修改内容完全取决于配置文件
// 但是，一般来说，改动的内容只涉及 label，assignees，state
// 而 title，body，milestone 不会改变
func IssueEdit(info Info, flow config.Flow) {
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
		hc := Comment{}
		hc.Login = info.Login
		// 这可能是一个修改重置时间的指令
		if flow.JobName == "reset" {
			if job, ok := config.Jobs[flow.JobName]; ok {
				resetDate, err := getResetDate(info.Owner, info.Repository, info.IssueNumber, flow.CurrentLabel, int(job.In), int(flow.Delay), flow.Name)
				if err != nil {
					fmt.Printf("get reset date for instruct failed. instruct: %v, err: %v\n", flow.Name, err.Error())
					return
				}
				hc.ResetDate = resetDate.In(time.Local).Format("2006-01-02")
			}
		}
		IssueComment(info, hc.HandComment(flow.SuccessFeedback))
	}
}

// 根据 flow 更新 info 中的 label
func updateLabel(req *gg.IssueRequest, info Info, flow config.Flow) {
	// 需要移除的 label
	remove := make(map[string]bool)
	for _, v := range flow.RemoveLabel {
		remove[v] = true
	}

	// 在添加下一阶段的 label
	// 对于想要保留的 label（目前会删除 current_label），只需要将其添加到 target_label 列表即可保留
	labels := make([]string, 0)
	// target label 总是会直接添加
	labels = append(labels, flow.TargetLabel...)
	for _, v := range info.Labels {
		// flow.RemoveLabel 的 label 会被忽略
		if _, ok := remove[v]; ok {
			continue
		}
		// 不在 flow.RemoveLabel 中的 label 会添加至新列表
		labels = append(labels, v)
	}
	req.Labels = &labels
}

// 根据 flow 更新 info 中的 assignees
func updateAssign(req *gg.IssueRequest, info Info, flow config.Flow) {
	defer func() {
		req.Assignees = &info.Assignees
	}()
	// todo 提出
	if flow.Name == "/accept" {
		info.Assignees = append(info.Assignees, info.Login)
	}
	switch flow.Mention {
	case "addition":
		info.Assignees = append(info.Assignees, info.Mention...)
	case "remove":
		removeAssign(req, info)
	default:
		// assignees 不增不减
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
