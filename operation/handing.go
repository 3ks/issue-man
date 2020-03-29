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
	fmt.Printf("do: %s, mention: %#v, payload: %#v\n", ins, mention, payload)

	// 基本信息
	info := GetInfo(payload)
	info.Mention = mention
	flow := config.Instructions[ins]

	// 权限检查
	if !CheckPermission(flow.Permission, info) {
		if flow.PermissionFeedback == "" {
			return
		}
		IssueComment(info, HandComment(flow.PermissionFeedback, info.Login, 0))
		return
	}

	// 标签（状态）检查
	if !CheckLabel(info.Labels, flow.CurrentLabel) {
		if flow.FailFeedback == "" {
			return
		}
		IssueComment(info, HandComment(flow.FailFeedback, info.Login, 0))
		return
	}

	// 数量检查
	if !CheckCount(info, flow.TargetLabel, flow.TargetLimit) {
		if flow.LimitFeedback == "" {
			return
		}
		IssueComment(info, HandComment(flow.LimitFeedback, info.Login, flow.TargetLimit))
		return
	}

	// 发送 Update Issue 请求（如果有的话）
	IssueEdit(info, flow)

	// 发送 Move Card 请求（如果有的话）
	CardMove(info, flow)
}
