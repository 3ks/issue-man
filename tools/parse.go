package tools

import (
	"github.com/google/uuid"
	"gopkg.in/go-playground/webhooks.v5/github"
	"issue-man/comm"
	"issue-man/global"
	"path"
	"regexp"
	"strconv"
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
func (p parseFunctions) Instruct(body string) (instructs map[string][]string) {
	instructs = make(map[string][]string)

	// 替换字符
	body = strings.ReplaceAll(body, "\r\n", "\n")
	// 按行分割
	s := strings.SplitN(body, "\n", -1)

	// 遍历行
	for _, v := range s {
		// 尝试解析指令
		is, peoples := p.parseInstruct(v)
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

// 解析指令，不导出
func (p parseFunctions) parseInstruct(s string) (is string, peoples []string) {
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

// PRNumberFromBody
// 从 body 内解析出 pr number
// PRNumberFromBody() 有一个对应的生成方法 BodyByPRNumberAndSha()
func (p parseFunctions) PRNumberFromBody(body string) (number int) {
	defer func() {
		if p := recover(); p != nil {
			global.Sugar.Panicw("bad body",
				"body", body,
				"panic", p)
			number = 0
		}
	}()
	firstLine := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")[0]
	number, _ = strconv.Atoi(path.Base(firstLine))
	return
}

// Info
// 从 IssueCommentPayload 里的一些信息
// 避免多次书写出现错误
func (p parseFunctions) Info(payload github.IssueCommentPayload) (info comm.Info) {
	defer func() {
		if p := recover(); p != nil {
			global.Sugar.Errorw("Info panic",
				"req_id", info.ReqID,
				"panic", p)
		}
	}()
	info.ReqID = uuid.New().String()
	info.Owner = payload.Repository.Owner.Login
	info.Repository = payload.Repository.Name

	info.Login = payload.Sender.Login

	info.IssueURL = payload.Issue.URL
	info.IssueNumber = int(payload.Issue.Number)
	info.Title = payload.Issue.Title
	info.Body = payload.Issue.Body
	info.Milestone = int(payload.Issue.Milestone.Number)
	info.State = payload.Issue.State

	info.Assignees = make([]string, len(payload.Issue.Assignees))
	info.Labels = make([]string, len(payload.Issue.Labels))
	for i := 0; i < len(payload.Issue.Assignees) || i < len(payload.Issue.Labels); i++ {
		if i < len(info.Assignees) {
			info.Assignees[i] = payload.Issue.Assignees[i].Login
		}
		if i < len(info.Labels) {
			info.Labels[i] = payload.Issue.Labels[i].Name
		}
	}
	return
}
