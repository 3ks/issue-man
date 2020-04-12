package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"issue-man/config"
	"net/http"
	"strings"
)

const (
	Top    = "top"
	Bottom = "bottom"
)

// 移动 issue 对应的 card（如果有的的话）
// todo 目前只能移动已存在的 card，没有添加的功能
func CardMove(info Info, flow config.Flow) {
	// 当前无 CurrentColumnID，不做操作
	if flow.CurrentColumnID == 0 {
		return
	}
	// 先获取 card 列表
	cards, resp, err := client.Get().Projects.ListProjectCards(context.TODO(), flow.CurrentColumnID, nil)
	if err != nil {
		fmt.Printf("list column card fail. err: %v\n", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("list column card maybe fail. status code: %v\n", resp.StatusCode)
		return
	}

	// 位置
	position := ""
	switch flow.TargetPosition {
	case Bottom:
		position = Bottom
	default:
		position = Top
	}

	// 遍历寻找 content_url 与当前 issue_url 相同的 card
	// 将其视为关联的 card，并将其移动至目标 column
	for _, v := range cards {
		if *v.ContentURL == info.IssueURL {
			req := &gg.ProjectCardMoveOptions{
				Position: position,
				ColumnID: flow.TargetColumnID,
			}
			resp, err := client.Get().Projects.MoveProjectCard(context.TODO(), *v.ID, req)
			if err != nil {
				fmt.Printf("move card fail. err: %v\n", err.Error())
				return
			}
			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("move card maybe fail. status code: %v\n", resp.StatusCode)
				return
			}
			if flow.ColumnFeedback != "" {
				IssueComment(info, strings.ReplaceAll(flow.ColumnFeedback, "@somebody", fmt.Sprintf("@%s", info.Login)))
			}
			break
		}
	}
}
