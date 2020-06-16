package global

import (
	"context"
	"github.com/google/go-github/v30/github"
	"net/http"
)

// 获取 Team 成员列表
// 通过 https://developer.github.com/v3/teams/members/#list-team-members  获取团队成员
// 通过 https://developer.github.com/webhooks/event-payloads/#membership Webhook 监听
// 组织成员，有人员变动时，调用该函数，重新加载成员列表
func LoadMaintainers() {
	// 分页器
	op := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	// 新的成员列表
	maintainers := make(map[string]bool)

	for {
		users, resp, err := Client.Teams.ListTeamMembersBySlug(context.Background(),
			Conf.Repository.Spec.Workspace.Owner,
			Conf.Repository.Spec.MaintainerTeam,
			op)
		if err != nil {
			Sugar.Errorw("load maintainer list",
				"call api", "failed",
				"err", err.Error(),
			)
			break
		}
		if resp.StatusCode != http.StatusOK {
			Sugar.Errorw("load maintainer list",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			break
		}

		// 存储数据
		for k := range users {
			maintainers[users[k].GetLogin()] = true
		}

		// 无更多数据，终止循环调用
		if len(users) < op.PerPage {
			break
		}
		// 下一页
		op.Page++
	}

	Lock.Lock()
	Maintainers = maintainers
	Lock.Unlock()

	Sugar.Infow("load maintainer list",
		"status", "done",
		"list", Maintainers)
}

// 获取 Members 成员列表
// 通过 https://developer.github.com/v3/orgs/members/#members-list 获取组织成员
// 通过 https://developer.github.com/webhooks/event-payloads/#organization Webhook 监听
// 组织成员，有人员变动时，调用该函数，重新加载成员列表。
// 当成员不存在时，也可以调用该函数，重新加载成员列表，以防止用户修改用户名的情况。
func LoadMembers() {
	// 分页器
	op := &github.ListMembersOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	// 新的成员列表
	members := make(map[string]bool)

	for {
		users, resp, err := Client.Organizations.ListMembers(context.Background(),
			Conf.Repository.Spec.Workspace.Owner,
			op)
		if err != nil {
			Sugar.Errorw("load member list",
				"call api", "failed",
				"err", err.Error(),
			)
			break
		}
		if resp.StatusCode != http.StatusOK {
			Sugar.Errorw("load member list",
				"call api", "unexpect status code",
				"status", resp.Status,
				"status code", resp.StatusCode,
				"response", resp.Body,
			)
			break
		}

		// 存储数据
		Lock.Lock()
		for k := range users {
			Members[users[k].GetLogin()] = true
		}
		Lock.Unlock()

		// 无更多数据，终止循环调用
		if len(users) < op.PerPage {
			break
		}
		// 下一页
		op.Page++
	}

	Lock.Lock()
	Members = members
	Lock.Unlock()

	Sugar.Infow("load member list",
		"status", "done",
		"list", Members)
}
