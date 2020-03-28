package operation

import (
	"context"
	"fmt"
	"issue-man/client"
	"issue-man/instruction"
)

const (
	Maintainer = "Maintainers"
	Self       = "self"
	Member     = "member"
	Anyone     = "anyone"
)

// 权限检查
// 行为：根据检查策略，对评论人进行权限检查
// 返回值为 true，则表示通过检测。
// 反之则表示未通过检测。
// 检查流程是，先检测配置文件是否配置了改项，如果配置了，用户是否满足该项的条件。
// 满足任一一个条件，则视为有权限。
func CheckPermission(permission []string, info Info) bool {
	// 未配置任何权限，则不允许操作
	if len(permission) == 0 {
		return false
	}

	// 权限 map
	ps := make(map[string]bool)
	for _, v := range permission {
		ps[v] = true
	}

	// 当前 assignees map
	as := make(map[string]bool)
	for _, v := range info.Assignees {
		as[v] = true
	}

	// Maintainers 可以操作
	if _, ok := ps[Maintainer]; ok {
		if _, ok := instruction.Maintainers[info.Login]; ok {
			return true
		}
	}

	// self 可以操作
	if _, ok := ps[Self]; ok {
		// 自身在当前 assignees 列表中
		if _, ok = as[info.Login]; ok {
			return true
		}
	}

	// member 可以操作
	if _, ok := ps[Member]; ok {
		if isMember(info) {
			return true
		}
	}

	// anyone 可以操作
	if _, ok := ps[Anyone]; ok {
		return true
	}

	return false
}

func isMember(info Info) bool {
	is, _, err := client.Get().Organizations.IsMember(context.TODO(), info.Owner, info.Login)
	if err != nil {
		fmt.Printf("query is member fail err: %v\n", err.Error())
		return false
	}

	return is
}
