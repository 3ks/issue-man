// init.go 对应的是 init 子命令。
// 一般仅在项目初始化时使用，后续创建应由主程序维护。
// 效果是根据上游仓库的内容和规则，在任务仓库创建对应的 issue 列表。
package cmd

import (
	"github.com/spf13/cobra"
	"issue-man/server"
)

var (
	initCmd *cobra.Command
)

func init() {
	// init
	initCmd = &cobra.Command{
		Use:   "start",
		Short: "初始化仓库内容。",
		Long:  `根据上游仓库内容和规则，初始化任务仓库内容。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化配置初始化服务相关的东西
			server.Init(loadAndInit())
		},
	}

	// 添加至 root 节点
	rootCmd.AddCommand(initCmd)

	// 解析参数
	initCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "GitHub Person Token.")
	initCmd.PersistentFlags().StringVarP(&c, "config", "c", "", "指定配置文件路径")
}
