package tools

// HasLabel
// 要求 label 的每一个元素都在 source 之中
func (v verifyFunctions) HasLabel(source *[]string, label ...string) bool {
	if len(label) == 0 {
		return true
	}
	if source == nil {
		return false
	}

	tmp := make(map[string]bool)
	for _, value := range label {
		tmp[value] = true
	}
	for _, value := range *source {
		if !tmp[value] {
			return false
		}
	}
	return true
}
