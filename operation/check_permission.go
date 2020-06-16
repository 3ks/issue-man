package operation

import (
	"issue-man/global"
)

const (
	Anyone     = "anyone"
	Assignees  = "assignees"
	Maintainer = "maintainers"
	Member     = "member"
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

	// 要求的权限 map
	ps := make(map[string]bool)
	for _, v := range permission {
		ps[v] = true
	}

	// 当前的 assignees map
	as := make(map[string]bool)
	for _, v := range info.Assignees {
		as[v] = true
	}

	// maintainer 可以操作
	if _, ok := ps[Maintainer]; ok {
		if global.Maintainers[info.Login] {
			return true
		}
	}

	// assigner 可以操作
	if _, ok := ps[Assignees]; ok {
		// 自身在当前 assignees 列表中
		if _, ok = as[info.Login]; ok {
			return true
		}
	}

	// member 可以操作
	if _, ok := ps[Member]; ok {
		if global.Members[info.Login] {
			return true
		}
		// 加载 member 列表，再检测一次
		global.LoadMembers()
		if global.Members[info.Login] {
			return true
		}
	}

	// anyone 可以操作
	if _, ok := ps[Anyone]; ok {
		return true
	}

	return false
}
