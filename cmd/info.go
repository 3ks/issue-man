package cmd

import (
	"github.com/spf13/cobra"
)

var (
	info *cobra.Command
)

func init() {
	// destroy
	info = &cobra.Command{
		Use:   "info",
		Short: "清空仓库内容。",
		Long:  `根据规则，清空任务仓库的内容。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化配置初始化服务相关的东西
			_ = loadAndInit()
		},
	}

	// 添加至 root 节点
	rootCmd.AddCommand(info)

	// 解析参数
	info.PersistentFlags().StringVarP(&token, "token", "t", "", "GitHub Person Token.")
	info.PersistentFlags().StringVarP(&c, "config", "c", "", "指定配置文件路径")
}
