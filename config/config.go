package config

import (
	"fmt"
	"path"
	"strings"
)

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
		CommitIssue    *int    `yaml:"commitIssue"`
		MaintainerTeam *string `yaml:"maintainerTeam"`
		Port           *string `yaml:"port"`
		LogLevel       *string `yaml:"logLevel"`
		Verbose        *bool   `yaml:"verbose"`
	} `yaml:"spec"`
}

// 创建 Issue 相关的配置
type IssueCreate struct {
	Base
	Spec struct {
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

func (i Include) OK(p string) bool {
	// 仅判断 .md 文件
	if path.Ext(p) != ".md" {
		return false
	}

	// 不包含前缀
	if !strings.HasPrefix(p, *i.Path) {
		return false
	}

	// 排除的子目录
	for _, v := range i.Exclude {
		if strings.HasPrefix(p, *v.Path) {
			return false
		}
	}
	return true
}

// Issue Comment 相关的配置
// 也就是指令相关的配置
// 创建 Issue 相关的配置
type IssueComment struct {
	Base
	Spec struct {
		Rules  *Rule   `yaml:"rules"`
		Action *Action `yaml:"action"`
	} `yaml:"spec"`
}

// 条件
type Rule struct {
	Instruct    *string   `yaml:"instruct"`
	Permissions []*string `yaml:"permissions"`
	Labels      []*string `yaml:"labels"`
	Assigners   []*string `yaml:"assigners"`
}

// 动作
type Action struct {
	AddLabels       []*string `yaml:"addLabels"`
	AddLabelsLimit  *int      `yaml:"addLabelsLimit"`
	RemoveLabels    []*string `yaml:"removeLabels"`
	AddAssigners    []*string `yaml:"addAssigners"`
	RemoveAssigners []*string `yaml:"removeAssigners"`
	State           *string   `json:"state"`
	SuccessFeedback *string   `yaml:"successFeedback"`
	FailFeedback    *string   `yaml:"failFeedback"`
}

// Job 定时任务相关的配置
// 也就是定时更新和状态检测相关的配置
// 同时，依赖 `创建 Issue` 的配置
type Job struct {
	Base
	Spec struct {
		// TODO
		//Rules  *Option `yaml:"rules"`
		//Action *Option `yaml:"action"`
	} `yaml:"spec"`
}
