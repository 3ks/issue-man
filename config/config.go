package config

import "issue-man/instruction"

type Config struct {
	// 完整的仓库名字，即 组织名+仓库名。如：servicemesher/istio-handbook
	FullRepositoryName string `mapstructure:"full_repository_name"`

	// Maintains 列表
	Maintains []string `mapstructure:"maintains"`

	// 通过配置文件定义任务流程
	Flows []instruction.Flow `mapstructure:"flows"`

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
