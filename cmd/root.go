package cmd

import (
	"fmt"
	"issue-man/server"
	"os"

	"github.com/spf13/cobra"
)


var (
	token          string
	maxAcceptIssue int
)
func init() {
	rootCmd.PersistentFlags().StringVarP(&token, "token", "f", "", "GitHub Person Token.")
	rootCmd.PersistentFlags().IntVarP(&maxAcceptIssue, "count", "c", 0, "Max accept issues by one human.")
	// todo 指定配置文件、输出目录
	// todo 以仓库为单位设置配置
	// todo 动态读取配置 watch?

}

var rootCmd = &cobra.Command{
	Use:   "spacer",
	Short: "格式化 markdown 文件。",
	Long:  `spacer 用于为你格式化 markdown 文件，spacer 非常灵活，每一条格式化规则都是可插拔的，并且您可以根据自己的情况，轻松对其进行扩展。`,
	Run: func(cmd *cobra.Command, args []string) {
		//_ = cmd.Usage()
		fd, _ := os.OpenFile("bot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		os.Stdout = fd
		os.Stderr = fd
		server.Start(token, maxAcceptIssue)
	},
	// todo 子命令，查看配置文件
	// todo 子命令，不仅是管理 issue 状态，考虑 issue 的生命周期管理？包括扫描已有 issue，去重，创建等。
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
