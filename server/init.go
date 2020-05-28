// init.go 对应 init 子命令的实现
// init 实现的是根据上游仓库和规则，在任务仓库创建初始化 issue
package server

import "issue-man/config"

// 注意：这个 Init 并不是传统的初始化函数！
// Init 根据上游仓库和规则，在任务仓库创建初始化 issue
func Init(conf config.Config) {

}
