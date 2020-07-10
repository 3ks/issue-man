package operation

import (
	gg "github.com/google/go-github/v30/github"
	"issue-man/comm"
	"issue-man/config"
	"issue-man/tools"
)

const (
	IssueOpen   = "open"
	IssueClosed = "closed"
	Commenter   = "@commenter"    // 评论者
	Mention     = "@mention"      // comment 提及的人
	AllAssignee = "@all-assignee" // 所有 assignees
)

// 对 issue 进行修改
// 具体修改内容完全取决于配置文件
// 但是，一般来说，改动的内容只涉及 label，assignees，state
// 而 title，body，milestone 不会改变
func issueEdit(info comm.Info, flow config.IssueComment) {
	// 一般不会变化的内容
	edit := &gg.IssueRequest{
		Title:     &info.Title,
		Body:      &info.Body,
		Milestone: &info.Milestone,
		State:     tools.Get.String(IssueOpen),
	}
	if edit.GetMilestone() == 0 {
		edit.Milestone = nil
	}
	// 是否关闭 issue
	if flow.Spec.Action.State == IssueClosed {
		edit.State = tools.Get.String(IssueClosed)
	}

	// 更新 label（如果有的话）
	updateLabel(edit, info, flow)

	// 更新 assignees（如果有的话）
	updateAssign(edit, info, flow)

	// 尝试调用更新接口
	_, err := tools.Issue.EditByIssueRequest(info.IssueNumber, edit)
	if err != nil {
		return
	}

	comment := comm.Comment{}
	comment.ReqID = info.ReqID
	comment.Login = info.Login
	comment.Assignees = info.Assignees
	// 这可能是一个修改重置时间的指令，解析其重置时间
	//if flow.JobName == "reset" {
	//	if Sync, ok := global.Jobs[flow.JobName]; ok {
	//		resetDate, err := getResetDate(info.Owner, info.Repository, info.IssueNumber, flow.CurrentLabel, int(Sync.In), int(flow.Delay), flow.Name)
	//		if err != nil {
	//			fmt.Printf("get reset date for instruct failed. instruct: %v, err: %v\n", flow.Name, err.Error())
	//			return
	//		}
	//		comment.ResetDate = resetDate.In(time.Local).Format("2006-01-02")
	//	}
	//}
	// 如果 feedback 为空不会做任何操作
	tools.Issue.Comment(info.IssueNumber, comment.HandComment(flow.Spec.Action.SuccessFeedback))
}

// 根据 flow 更新 info 中的 label
func updateLabel(req *gg.IssueRequest, info comm.Info, flow config.IssueComment) {
	req.Labels = tools.Convert.SliceAdd(tools.Convert.SliceRemove(tools.Get.Strings(info.Labels), flow.Spec.Action.RemoveLabels...), flow.Spec.Action.AddLabels...)
}

// 根据 flow 更新 info 中的 assignees
func updateAssign(req *gg.IssueRequest, info comm.Info, flow config.IssueComment) {
	// 将当前 assignees 列表转为 map，方便操作
	assignMap := tools.Convert.StringToMap(info.Assignees)
	// defer 将最终 assignees 列表转为 slice 并赋值给 req
	defer func() {
		req.Assignees = tools.Convert.MapToString(assignMap)
	}()

	// 添加的 assigner
	for _, v := range flow.Spec.Action.AddAssignees {
		switch v {
		case Commenter:
			assignMap[info.Login] = true
		case Mention:
			for _, v := range info.Mention {
				assignMap[v] = true
			}
		}
	}

	// 移除的 assigner
	for _, v := range flow.Spec.Action.RemoveAssignees {
		switch v {
		case Commenter:
			delete(assignMap, info.Login)
		case Mention:
			for _, v := range info.Mention {
				delete(assignMap, v)
			}
		case AllAssignee:
			assignMap = make(map[string]bool)
			return
		}
	}
}
