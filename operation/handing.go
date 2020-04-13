package operation

import (
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/config"
)

// is 是一个指令 map，其中：
// key 为指令名
// value 为提及人员，可能为空
func IssueHanding(payload github.IssueCommentPayload, is map[string][]string) {
	for k, v := range is {
		if _, ok := config.Instructions[k]; ok {
			do(k, v, payload)
		} else {
			fmt.Printf("unkown instruction: %s, mention: %#v\n", k, v)
		}
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
func do(ins string, mention []string, payload github.IssueCommentPayload) {
	// 基本信息
	info := GetInfo(payload)
	info.Mention = mention
	flow := config.Instructions[ins]

	fmt.Printf("do: %s, mention: %#v, info: %#v\n", ins, mention, info)

	// 权限检查
	if !CheckPermission(flow.Permission, info) {
		fmt.Printf("check permission fail. require: %v\n", flow.Permission)
		if flow.PermissionFeedback == "" {
			return
		}
		hc := Comment{}
		hc.Login = info.Login
		IssueComment(info, hc.HandComment(flow.PermissionFeedback))
		return
	}

	// 标签（状态）检查
	if !CheckLabel(info.Labels, flow.CurrentLabel) {
		fmt.Printf("check label fail. require: %v\n", flow.CurrentLabel)
		if flow.FailFeedback == "" {
			return
		}
		hc := Comment{}
		hc.Login = info.Login
		IssueComment(info, hc.HandComment(flow.PermissionFeedback))
		return
	}

	// 数量检查
	if !CheckCount(info, flow.TargetLabel, flow.TargetLimit) {
		fmt.Printf("check count fail. require: %v\n", flow.TargetLimit)
		if flow.LimitFeedback == "" {
			return
		}
		hc := Comment{}
		hc.Login = info.Login
		hc.Count = flow.TargetLimit
		IssueComment(info, hc.HandComment(flow.PermissionFeedback))
		return
	}

	// 发送 Update Issue 请求（如果有的话）
	IssueEdit(info, flow)

	// 发送 Move Card 请求（如果有的话）
	CardMove(info, flow)
}
