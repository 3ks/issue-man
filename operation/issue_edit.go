package operation

import (
	"context"
	gg "github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"net/http"
)

const (
	IssueOpen   = "open"
	IssueClosed = "closed"
	AsCommenter = "@commenter" // assignee 或 unassignee 评论者
	AsMention   = "@mention"   // assignee 或 unassignee 提及的人
)

// 对 issue 进行修改
// 具体修改内容完全取决于配置文件
// 但是，一般来说，改动的内容只涉及 label，assignees，state
// 而 title，body，milestone 不会改变
func IssueEdit(info Info, flow config.IssueComment) {
	// 一般不会变化的内容
	req := &gg.IssueRequest{
		Title:     &info.Title,
		Body:      &info.Body,
		Milestone: &info.Milestone,
	}

	closeIssue := IssueOpen
	// 是否关闭 issue
	if flow.Spec.Action.State == IssueClosed {
		closeIssue = IssueClosed
	}
	req.State = &closeIssue

	// 更新 label（如果有的话）
	updateLabel(req, info, flow)

	// 更新 assignees（如果有的话）
	updateAssign(req, info, flow)

	// 尝试调用更新接口
	_, resp, err := global.Client.Issues.Edit(context.TODO(), info.Owner, info.Repository, info.IssueNumber, req)
	if err != nil {
		global.Sugar.Errorw("IssueEdit",
			"req_id", info.ReqID,
			"step", "call api",
			"info", info,
			"req", req,
			"err", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("CheckCount",
			"req_id", info.ReqID,
			"step", "parse response",
			"info", info,
			"req", req,
			"statusCode", resp.StatusCode,
			"body", string(body))
		return
	}

	// 创建文本提示
	if flow.Spec.Action.SuccessFeedback != "" {
		hc := Comment{}
		hc.Login = info.Login
		// 这可能是一个修改重置时间的指令，解析其重置时间
		//if flow.JobName == "reset" {
		//	if job, ok := global.Jobs[flow.JobName]; ok {
		//		resetDate, err := getResetDate(info.Owner, info.Repository, info.IssueNumber, flow.CurrentLabel, int(job.In), int(flow.Delay), flow.Name)
		//		if err != nil {
		//			fmt.Printf("get reset date for instruct failed. instruct: %v, err: %v\n", flow.Name, err.Error())
		//			return
		//		}
		//		hc.ResetDate = resetDate.In(time.Local).Format("2006-01-02")
		//	}
		//}
		IssueComment(info, hc.HandComment(flow.Spec.Action.SuccessFeedback))
	}
}

// 根据 flow 更新 info 中的 label
func updateLabel(req *gg.IssueRequest, info Info, flow config.IssueComment) {
	// 需要移除的 label
	remove := make(map[string]bool)
	for _, v := range flow.Spec.Action.RemoveLabels {
		remove[v] = true
	}

	// 添加下一阶段的 label
	labels := make([]string, 0)
	// target label 总是会直接添加
	for _, v := range flow.Spec.Action.AddLabels {
		labels = append(labels, v)
	}

	// 遍历目前存在的 label
	// 对于要求移除的 label，则不添加
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
// 一般成员只能操作自己，maintainer 才能操作提及的人员。
func updateAssign(req *gg.IssueRequest, info Info, flow config.IssueComment) {
	// 将当前 assign 列表转为 map，方便操作
	assign := make(map[string]bool)
	for _, v := range info.Assignees {
		assign[v] = true
	}

	defer func() {
		// 将最终 assign 列表转为 slice 并复制给 req
		tmp := make([]string, len(assign))
		count := 0
		for k := range assign {
			tmp[count] = k
			count++
		}
		req.Assignees = &tmp
	}()

	// 添加的 assigner
	for _, v := range flow.Spec.Action.AddAssigners {
		switch v {
		case AsCommenter:
			assign[info.Login] = true
		case AsMention:
			// 添加，仅 maintainer 可以 assign 提及的人
			// 普通成员只能通过指令，由系统自动 assign（如 accept），不能直接指定。
			if global.Maintainers[info.Login] {
				for _, v := range info.Mention {
					assign[v] = true
				}
			}
		}
	}

	// 移除的 assigner
	for _, v := range flow.Spec.Action.RemoveAssigners {
		switch v {
		case AsCommenter:
			info.Assignees = append(info.Assignees, info.Login)
			delete(assign, info.Login)
		case AsMention:
			// 移除，maintainer 可以 unassign 别人
			if global.Maintainers[info.Login] {
				for _, v := range info.Mention {
					delete(assign, v)
				}
			} else {
				for _, v := range info.Mention {
					// 普通成员仅可以移除 unassign 自己
					if v == info.Login {
						delete(assign, v)
					}
				}
			}
		}
	}
}
