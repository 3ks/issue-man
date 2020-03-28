package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"issue-man/instruction"
	"net/http"
)

// 移动 issue 对应的 card（如果有的的话）
func CardMove(info Info, flow instruction.Flow) {
	client.Get().Teams.ListTeamMembersByID()
	cards, resp, err := client.Get().Projects.ListProjectCards(context.TODO(), flow.CurrentColumnID, nil)
	if err != nil {
		fmt.Printf("list column card fail. err: %v\n", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("list column card maybe fail. status code: %v\n", resp.StatusCode)
		return
	}
	for _, v := range cards {
		if *v.ContentURL == info.IssueURL {
			req := &gg.ProjectCardMoveOptions{
				Position: flow.TargetPosition,
				ColumnID: flow.TargetColumnID,
			}
			resp, err := client.Get().Projects.MoveProjectCard(context.TODO(), *v.ID, req)
			if err != nil {
				fmt.Printf("move card fail. err: %v\n", err.Error())
				return
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("move card maybe fail. status code: %v\n", resp.StatusCode)
				return
			}
			break
		}
	}
}
