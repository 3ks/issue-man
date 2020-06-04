// start.go 对应 start 子命令，表示启动程序。
package cmd

import (
	"github.com/spf13/cobra"
	"issue-man/server"
)

var (
	startCmd *cobra.Command
)

func init() {
	// start
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "开始运行。",
		Long:  `开始运行 Issue Man。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化配置初始化服务相关的东西
			server.Start(loadAndInit())
		},
	}

	// 添加至 root 节点
	rootCmd.AddCommand(startCmd)

	// 解析参数
	startCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "GitHub Person Token.")
	startCmd.PersistentFlags().StringVarP(&c, "config", "c", "", "指定配置文件路径")
}
