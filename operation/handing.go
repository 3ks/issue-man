package operation

import (
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/instruction"
	"strings"
)

// is 是一个指令 map，其中：
// key 为指令名
// value 为提及人员，可能为空
func IssueHanding(payload github.IssueCommentPayload, is map[string][]string) {
	for k, v := range is {
		if _, ok := instruction.Instructions[k]; ok {
			do(k, v, payload)
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
//
// 在检查过程中，随时可能会 comment，并 return，
// 这取决于 issue 的实际情况和流程定义
func do(instruction string, mention []string, payload github.IssueCommentPayload) {
	// todo 日志，配置
	fmt.Printf("do: %s, mention: %s, payload: %#v\n", instruction, mention, payload)

	// 基本信息
	info := GetInfo(payload)
	info.Mention = mention
	flow := instruction.Instructions[instruction]
	commentBody := ""

	// 权限检查
	if !CheckPermission(flow.Permission, info) {
		if flow.PermissionFeedback == "" {
			return
		}
		commentBody = strings.ReplaceAll(flow.PermissionFeedback, "@somebody", info.Login)
		IssueComment(info, commentBody)
		return
	}

	// 标签（状态）检查
	if !CheckLabel(info.Labels, flow.CurrentLabel) {
		if flow.FailFeedback == "" {
			return
		}
		commentBody = strings.ReplaceAll(flow.PermissionFeedback, "@somebody", info.Login)
		IssueComment(info, commentBody)
		return
	}

	// 数量检查
	if !CheckCount(info, flow.TargetLabel, flow.TargetLimit) {
		if flow.FailFeedback == "" {
			return
		}
		commentBody = strings.ReplaceAll(flow.PermissionFeedback, "@somebody", info.Login)
		IssueComment(info, commentBody)
		return
	}

	// 发送 Update Issue 请求（如果有的话）
	IssueEdit(info, flow)

	// 发送 Move Card 请求（如果有的话）
	CardMove(info, flow)
}
