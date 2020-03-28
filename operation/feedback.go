package operation

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Somebody = "@somebody"
	Count    = "@count"
)

// 替换文本提示里的特殊字符
// 后面可能会再增加一些特殊字符
func HandComment(text string, login string, count int) string {
	text = strings.ReplaceAll(text, Somebody, fmt.Sprintf("@%s", login))
	text = strings.ReplaceAll(text, Count, strconv.Itoa(count))
	return text
}
