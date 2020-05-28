package config

import "fmt"

type Config struct {
	Repository    *Repository     `yaml:"repository"`
	IssueCreate   *IssueCreate    `yaml:"issue_create"`
	IssueComments []*IssueComment `yaml:"issue_comment"`
	Jobs          []*Job          `yaml:"jobs"`
}

type Base struct {
	ApiVersion string `yaml:"api_version"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

type Selector struct {
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
}

func (s Selector) GetFullName() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repository)
}

// 仓库及一些全局相关的配置
type Repository struct {
	Base
	Spec struct {
		Workspace Selector `yaml:"workspace"`
		Upstream  Selector `yaml:"upstream"`
		// TODO 手动创建一个特殊的 issue，用于存储 commit 指针。
		CommitIssue *int `yaml:"commit_issue"`
		// TODO 动态维护 maintainers 列表和 member 列表
		// 通过 https://developer.github.com/v3/teams/members/#list-team-members  获取 maintainer 成员
		// 通过 https://developer.github.com/webhooks/event-payloads/#membership Webhook 监听团队成员变动并更新
		// Member 同理，
		// 通过 https://developer.github.com/v3/orgs/members/#members-list 获取组织成员
		// 通过 https://developer.github.com/webhooks/event-payloads/#organization Webhook 监听组织成员并更新
		MaintainerTeam *int    `yaml:"maintainer_team"`
		Port           *string `yaml:"port"`
		LogLevel       *string `yaml:"log_level"`
		Verbose        *bool   `yaml:"verbose"`
	} `yaml:"spec"`
}

// 创建 Issue 相关的配置
type IssueCreate struct {
	Base
	Spec struct {
		Title string `yaml:"title"`
		// Title 要跳过的段数
		TitleSkip *int `yaml:"titleSkip"`
		// 选择风格
		Body      string     `yaml:"body"`
		Labels    []*string  `yaml:"labels"`
		Assignees []*string  `yaml:"assignees"`
		Milestone *int       `yaml:"milestone"`
		Includes  []*Include `yaml:"includes"`
	} `yaml:"spec"`
}

type Include struct {
	Path    *string    `yaml:"path"`
	Labels  []*string  `yaml:"labels"`
	Exclude []*Include `yaml:"exclude"`
}

// Issue Comment 相关的配置
// 也就是指令相关的配置
// 创建 Issue 相关的配置
type IssueComment struct {
	Base
	Spec struct {
		Rules  *Option `yaml:"rules"`
		Action *Option `yaml:"action"`
	} `yaml:"spec"`
}

// 选项
type Option struct {
	Instruct        *string   `yaml:"instruct"`
	Permissions     []*string `yaml:"permissions"`
	In              *int      `yaml:"in"`
	Labels          []*string `yaml:"labels"`
	AddLabels       []*string `yaml:"add_labels"`
	AddLabelsLimit  *int      `yaml:"add_labels_limit"`
	RemoveLabels    []*string `yaml:"remove_labels"`
	Assigners       []*string `yaml:"assigners"`
	AddAssigners    []*string `yaml:"add_assigners"`
	RemoveAssigners []*string `yaml:"remove_assigners"`
	SuccessFeedback *string   `yaml:"success_feedback"`
	FailFeedback    *string   `yaml:"fail_feedback"`
}

// Job 定时任务相关的配置
// 也就是定时更新和状态检测相关的配置
// 创建 Issue 相关的配置
type Job struct {
	Base
	Spec struct {
		Rules  *Option `yaml:"rules"`
		Action *Option `yaml:"action"`
	} `yaml:"spec"`
}
