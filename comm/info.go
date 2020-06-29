package comm

// 存储 IssueCommentPayload 里的一些信息
// 基本是目前进行各种操作需要用到的信息
type Info struct {
	// 仓库信息
	Owner      string
	Repository string

	// 评论人信息
	Login string
	// 评论提及到的人
	Mention []string

	// Issue 目前的信息
	IssueURL    string
	IssueNumber int
	Title       string
	Body        string
	Milestone   int
	State       string
	Assignees   []string
	Labels      []string

	// 一个指令的 UUID
	ReqID string
}
