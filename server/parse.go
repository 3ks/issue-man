package server

import (
	"regexp"
	"strings"
)

// 解析指令
// 尝试根据给定的文本解析指令
// 支持一行 @ 多人，例如：/cc @noone  @someone
// 也支持多行 @ 多人，例如：
//		/cc @noone
//		/cc @someone
// 是否支持指令，是否支持指令 @ 某人，决定权交给 operation 处理。
// 处理后的指令类似于： /accept  /pushed 等
// 处理后的 @某人，类似于：@someone
// 返回值结构为 key-value，其中：
// key 为指令名
// value 为提及员，可能为空
func ParseInstruct(body string) (instructs map[string][]string) {
	instructs = make(map[string][]string)

	// 按行分割
	s := strings.SplitN(body, "\n", -1)

	// 遍历行
	for _, v := range s {
		// 尝试解析指令
		is, peoples := parse(v)
		if is == "" {
			continue
		}
		// 初始化数组
		if instructs[is] == nil {
			instructs[is] = make([]string, 0)
		}
		// 添加相关人员
		instructs[is] = append(instructs[is], peoples...)
	}
	return
}

// 解析指令
func parse(s string) (is string, peoples []string) {
	s += " "
	strings.TrimLeft(s, " ")
	//expName:=regexp.MustCompile("(?<=^/).*?(?= )")
	//expPeople:=regexp.MustCompile("(?<=@).*?(?= )")
	is = regexp.MustCompile("^/.*? ").FindString(s)
	if is == "" {
		return
	}
	is = strings.TrimSpace(is)

	peoples = regexp.MustCompile("^@.*? ").FindAllString(s, -1)
	if len(peoples) == 0 {
		return
	}
	for k, v := range peoples {
		peoples[k] = strings.TrimSpace(v)
	}
	return
}
