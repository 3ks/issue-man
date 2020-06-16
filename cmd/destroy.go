// destroy.go 对应的是 destroy 子命令
// 效果是根据规则（label）清空任务仓库的所有 issue。
package cmd

import (
	"github.com/spf13/cobra"
	"issue-man/server"
)

var (
	destroyCmd *cobra.Command
)

func init() {
	// destroy
	destroyCmd = &cobra.Command{
		Use:   "destroy",
		Short: "清空仓库内容。",
		Long:  `根据规则，清空任务仓库的内容。`,
		Run: func(cmd *cobra.Command, args []string) {
			// 初始化配置初始化服务相关的东西
			server.Destroy(loadAndInit())
		},
	}

	// 添加至 root 节点
	rootCmd.AddCommand(destroyCmd)

	// 解析参数
	destroyCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "GitHub Person Token.")
	destroyCmd.PersistentFlags().StringVarP(&c, "config", "c", "", "指定配置文件路径")
}
