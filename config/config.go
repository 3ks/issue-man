package config

import (
	"fmt"
	"path"
	"strings"
)

type Config struct {
	Repository    Repository     `yaml:"repository"`
	IssueCreate   IssueCreate    `yaml:"issue_create"`
	IssueComments []IssueComment `yaml:"issue_comment"`
	Jobs          []Job          `yaml:"jobs"`
}

type Base struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

type Selector struct {
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Site       string `yaml:"site"`
	Branch     string `yaml:"branch"` // 分支
}

// 拼装 owner 和 repository
func (s Selector) GetFullName() string {
	return fmt.Sprintf("%s/%s", s.Owner, s.Repository)
}

// 仓库及一些全局相关的配置
type Repository struct {
	Base
	Spec struct {
		Source    Selector `yaml:"source"`    // 源库
		Translate Selector `yaml:"translate"` // 翻译库
		Workspace struct {
			Owner          string `yaml:"owner"`
			Repository     string `yaml:"repository"`
			MaintainerTeam string `yaml:"maintainerTeam"`
			Detection      struct {
				Enable  bool   `yaml:"enable"` // 是否检测同步更新 issue
				At      string `yaml:"at"`
				PRIssue int    `yaml:"prIssue"`
				// Comment Need Label
				NeedLabel       []string `yaml:"needLabel"`
				AddLabel        []string `yaml:"addLabel"`
				RemoveLabel     []string `yaml:"removeLabel"`
				DeprecatedLabel []string `yaml:"deprecatedLabel"`
			} `yaml:"detection"`
			Labels []struct {
				Name        string `yaml:"name"`
				Description string `yaml:"description"`
			} `yaml:"labels"` // 初始化时，自动创建的 label
		} `yaml:"workspace"` // 工作库
		Port     string `yaml:"port"`
		LogLevel string `yaml:"logLevel"`
		Verbose  bool   `yaml:"verbose"`
	} `yaml:"spec"`
}

// 创建 Issue 相关的配置
type IssueCreate struct {
	Base
	Spec struct {
		Prefix string `yaml:"prefix"`
		// 默认为 false，即默认会在 title 里移除 prefix 的部分
		SaveTitlePrefix bool     `yaml:"saveTitlePrefix"`
		FileType        []string `yaml:"fileType"`
		Labels          []string `yaml:"labels"`
		Assignees       []string `yaml:"assignees"`
		Milestone       int      `yaml:"milestone"`
		// 分类依据，可选值为 directory、file，默认为 directory
		GroupBy  string    `yaml:"groupBy"`
		Includes []Include `yaml:"includes"`
	} `yaml:"spec"`
}

// 验证文件是否为需要处理的文件格式
func (i IssueCreate) SupportType(file string) bool {
	// 前缀不匹配
	if !strings.HasPrefix(file, i.Spec.Prefix) {
		return false
	}

	ext := strings.ReplaceAll(path.Ext(file), ".", "")
	for _, v := range i.Spec.FileType {
		// 后缀匹配
		if v == ext {
			return true
		}
	}
	return false
}

// 判断是否处理该文件
// 如果处理，则返回其匹配的相关信息
func (i IssueCreate) SupportFile(filename string) (Include, bool) {
	// 仅处理支持的文件格式
	if !i.SupportType(filename) {
		return Include{}, false
	}

	for _, include := range i.Spec.Includes {
		// 包含关键字
		if strings.Contains(filename, include.Path) {
			// 如果不包含任一排除的子目录，则符合条件
			if !i.hasExclude(filename, include.Exclude) {
				return include, true
			}
		}
	}

	// 无匹配 include
	return Include{}, false
}

func (i IssueCreate) hasExclude(filename string, excludes []Include) bool {
	for _, v := range excludes {
		if strings.Contains(filename, v.Path) {
			return true
		}
	}
	return false
}

type Include struct {
	Path string `yaml:"path"`

	// 对于这一类文件，将 title 强制重写为配置文件指定的内容
	// 并且不显示 website 地址和 commit 历史界面
	Title string `yaml:"title"`

	// 分类依据，可选值为 directory、file，默认为 directory
	GroupBy string    `yaml:"groupBy"`
	Labels  []string  `yaml:"labels"`
	Exclude []Include `yaml:"exclude"`
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
	Instruct           string   `yaml:"instruct"`
	Permissions        []string `yaml:"permissions"`
	PermissionFeedback string   `yaml:"permissionFeedback"`
	Labels             []string `yaml:"labels"`
	LabelFeedback      string   `yaml:"labelFeedback"`
	Assignees          []string `yaml:"assignees"`
	AssignerFeedback   string   `yaml:"assignerFeedback"`
}

// 动作
type Action struct {
	AddLabels          []string `yaml:"addLabels"`
	AddLabelsLimit     int      `yaml:"addLabelsLimit"`
	LabelLimitFeedback string   `yaml:"labelLimitFeedback"`
	RemoveLabels       []string `yaml:"removeLabels"`
	AddAssignees       []string `yaml:"addAssignees"`
	RemoveAssignees    []string `yaml:"removeAssignees"`
	State              string   `yaml:"state"`
	SuccessFeedback    string   `yaml:"successFeedback"`
	FailFeedback       string   `yaml:"failFeedback"`
}

// Job 定时任务相关的配置
// 也就是定时更新和状态检测相关的配置
// 同时，依赖 `创建 Issue` 的配置
type Job struct {
	Base
	Spec struct {
		In           int      `yaml:"in"`
		Labels       []string `yaml:"labels"`
		AddLabels    []string `yaml:"addLabels"`
		RemoveLabels []string `yaml:"removeLabels"`
	} `yaml:"spec"`
}
