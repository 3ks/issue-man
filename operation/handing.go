package operation

import (
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/global"
)

// is 是一个指令 map，其中：
// key 为指令名
// value 为提及人员，可能为空
func IssueHanding(payload github.IssueCommentPayload, instructs map[string][]string) {
	for instruct, mention := range instructs {
		if _, ok := global.Instructions[instruct]; !ok {
			global.Sugar.Errorw("unknown instruction",
				"instruction", instruct,
				"mention", mention)
			continue
		}
		do(instruct, mention, payload)
	}
}

// 执行流程如下：
// 权限检查
// 状态检查
// 数量检查
// 拼装数据
// 发送 Update Issue 请求
// 发送 Move Card 请求（如果有的话）
//-----------------------------------------
// 在检查过程中，随时可能会 comment，并 return
// 这取决于 issue 的实际情况和流程定义
func do(instruct string, mention []string, payload github.IssueCommentPayload) {
	// 基本信息
	info := GetInfo(payload)
	info.Mention = mention
	flow := global.Instructions[instruct]

	global.Sugar.Debugw("do instruct",
		"req_id", info.ReqID,
		"instruct", instruct,
		"mention", mention,
		"info", info)

	// 权限检查
	if !CheckPermission(flow.Spec.Rules.Permissions, info) {
		global.Sugar.Infow("do instruct",
			"req_id", info.ReqID,
			"step", "CheckPermission",
			"status", "fail",
			"info", info,
			"require", flow.Spec.Rules.Permissions)
		if flow.Spec.Rules.PermissionFeedback == nil {
			return
		}
		hc := Comment{
			Login: info.Login,
			ReqID: info.ReqID,
		}
		IssueComment(info, hc.HandComment(*flow.Spec.Rules.PermissionFeedback))
		return
	}

	// 标签（状态）检查
	if !CheckLabel(flow.Spec.Rules.Labels, info.Labels) {
		global.Sugar.Infow("do instruct",
			"req_id", info.ReqID,
			"step", "CheckLabel",
			"status", "fail",
			"info", info,
			"require", flow.Spec.Rules.Labels)
		if flow.Spec.Rules.LabelFeedback == nil {
			return
		}
		hc := Comment{
			Login: info.Login,
			ReqID: info.ReqID,
		}
		IssueComment(info, hc.HandComment(*flow.Spec.Rules.LabelFeedback))
		return
	}

	// 数量检查
	if !CheckCount(info, flow.Spec.Action) {
		fmt.Printf("check count fail. require: %v\n", flow.TargetLimit)
		global.Sugar.Infow("do instruct",
			"req_id", info.ReqID,
			"step", "CheckCount",
			"status", "fail",
			"requireCount", flow.Spec.Action.AddLabelsLimit)
		if flow.Spec.Action.LabelLimitFeedback == nil {
			return
		}
		hc := Comment{
			Login: info.Login,
			ReqID: info.ReqID,
		}
		IssueComment(info, hc.HandComment(*flow.Spec.Action.LabelLimitFeedback))
		return
	}

	// 发送 Update Issue 请求（如果有的话）
	IssueEdit(info, *flow)

	// 发送 Move Card 请求（如果有的话）
	//CardMove(info, flow)
}
