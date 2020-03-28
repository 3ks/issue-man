package operation

// 标签（状态）检查
// 行为：require 中的每一个值是否都存在于 now 中。
// 返回值为 true，则表示通过检测。
// 反之则表示未通过检测
func CheckLabel(now, require []string) bool {
	m := make(map[string]bool)
	for _, v := range now {
		m[v] = true
	}

	for _, v := range require {
		if !m[v] {
			return false
		}
	}
	return true
}
