package instruction

// 指令及其对应的 Flow
// 每个指令对应一个 Flow
// 支持哪些指令，取决于配置文件内容
// Flow 的工作流程，取决于配置文件内容
// Flow 的行为及处理逻辑，取决于配置文件
var Instructions map[string]Flow

// Maintainers
//
var Maintainers map[string]bool

// 指令
const (
	IAccept   = "/accept"
	IPush     = "/pushed"
	IMerged   = "/merged"
	IAssign   = "/assign"
	IUnassign = "/unassign"
)

// 状态
const (
	SPending   = "status/spending"
	SWaiting   = "status/waiting-for-pr"
	SReviewing = "status/reviewing"
	SFinish    = "status/merged"
)
