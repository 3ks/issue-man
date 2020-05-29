// init.go 对应 init 子命令的实现
// init 实现的是根据上游仓库和规则，在任务仓库创建初始化 issue
package server

import (
	"issue-man/config"
	"issue-man/global"
)

// 注意：这个 Init 并不是传统的初始化函数！
// Init 根据上游仓库和 IssueCreate 规则，在任务仓库创建初始化 issue
// 获取 path 获取文件，
// 1. 包含 _index 开头的文件的目录，创建统一的 issue（但会继续遍历相关子目录），由 maintainer 统一管理。
// 3. 以包含 .md 文件的目录为单位，创建 issue（即一个目录可能包含多个 .md 文件）
func Init(conf config.Config) {
	global.Client.Repositories.GetContents()
}
