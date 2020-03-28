package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "issue-man",
	Short: "issue 生命周期的机器人",
	Long:  `issue-man 一个用于管理 GitHub Issue 生命周期的机器人。`,
	Run: func(cmd *cobra.Command, args []string) {
		if token == "" {
			_ = cmd.Usage()
			return
		}
		start()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
