// 用于 handbook issue 任务
// /accept
//		判断状态是否为 /pending，
//			如果是，则将该 issue 分配给评论人，并修改 状态的 translating，并回复
// 			否则，回复该 issue 不可 accept
// /pushed
//		判断状态是否为 /waiting-for-pr，且评论人为被分配人
//			如果是，则将该 issue 状态修改为 reviewing
//			否则，回复提示
// / merged
//		判断状态是否为 /reviewing，且评论人为被分配人
//			如果是，则将该 issue 状态修改为 merged，并关闭该 issue
//			否则，回复提示
package main

import "issue-man/cmd"

func main() {
	cmd.Execute()
}

