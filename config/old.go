package config

type Config2 struct {
	// 完整的仓库名字，即 组织名+仓库名。如：servicemesher/istio-handbook
	FullRepositoryName string `mapstructure:"full_repository_name"`

	// 是否在相关 issue 内 comment 错误原因
	Verbose bool `mapstructure:"verbose"`

	// Maintains 列表
	Maintainers []string `mapstructure:"maintainers"`

	// Job 配置
	// 检测时间
	DetectionAt string `mapstructure:"detection_at"`
	// 周末放个假
	SkipWeekend bool `mapstructure:"skip_weekend"`

	// 通过配置文件定义任务流程
	Flows []Flow `mapstructure:"flows"`

	// 任务
	Jobs []Job `mapstructure:"jobs"`

	// 其它设置
	// 监听端口
	Port string `mapstructure:"port"`
	// 日志目录，默认为 ./log
	LogDir string `mapstructure:"log_dir"`
	// 日志文件，默认为 issue-man.log（位于 LogDir 下）
	LogFile string `mapstructure:"log_file"`
	// 标准输出文件，默认为 issue-man.std.log（位于 LogDir 下）
	StdOutFile string `mapstructure:"std_out_file"`
}

// Job
type Job2 struct {
	Name            string   `mapstructure:"name"`
	In              int64    `mapstructure:"in"`
	InstructName    string   `mapstructure:"instruct_name"`
	Labels          []string `mapstructure:"labels"`
	RemoveLabels    []string `mapstructure:"remove_labels"`
	TargetLabels    []string `mapstructure:"target_label"`
	AssigneesPolicy string   `mapstructure:"assignees_policy"`
	CurrentColumnID int64    `mapstructure:"current_column_id"`
	TargetColumnID  int64    `mapstructure:"target_column_id"`
	TargetPosition  string   `mapstructure:"target_position"`
	WarnFeedback    string   `mapstructure:"warn_feedback"`
	Feedback        string   `mapstructure:"feedback"`
}
