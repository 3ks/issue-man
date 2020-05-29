// comm.go 包含了各个子命令通用的一些遍历
// 如：配置、日志等
package global

import (
	"context"
	c "github.com/google/go-github/v30/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"issue-man/config"
	"sync"
)

// 各种全局对象
var (
	// 全局锁
	Lock sync.Mutex

	// 配置对象
	Conf *config.Config

	// Client Client
	Client *c.Client

	// 日志对象
	Sugar *zap.SugaredLogger

	// 任务仓库 maintainer 列表
	// 判断某个用户是否为 maintainer，直接使用变量，根据返回值即可判断
	// 例如：result:=global.Maintainers["gorda"]，然后判断返回值即可。
	Maintainers = make(map[string]bool)

	// 组织成员
	Members = make(map[string]bool)

	// 支持的指令列表
	// 指令及其对应的 Flow
	// 每个指令对应一个 Flow
	// 支持哪些指令，取决于配置文件内容
	// Flow 的工作流程，取决于配置文件内容
	// Flow 的行为及处理逻辑，取决于配置文件
	Instructions = make(map[string]*config.IssueComment)

	// 需要执行的任务列表
	Jobs = make(map[string]*config.Job)
)

// 根据配置初始化一些内容。
// 根据 Token 获得组织信息，并初始化一些信息。
func Init(token string, conf *config.Config) {
	Conf = conf
	// 生产环境日志
	if *Conf.Repository.Spec.LogLevel == "pro" {
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err.Error())
		}
		Sugar = logger.Sugar()
		Sugar.Infow("init logger", "level", "production")
	} else {
		// 开发环境日志
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err.Error())
		}
		Sugar = logger.Sugar()
		Sugar.Infow("init logger", "level", "development")
	}

	// 从配置文件读取指令列表
	for _, v := range Conf.IssueComments {
		Instructions[v.Metadata.Name] = v
	}
	Sugar.Infow("load instructs",
		"done", Instructions)

	// 从配置文件读取 Job 列表
	for _, v := range Conf.Jobs {
		if v.Spec.Rules.In == nil || *v.Spec.Rules.In < 0 {
			continue
		}
		Jobs[v.Metadata.Name] = v
	}
	Sugar.Infow("load jobs",
		"done", Instructions)

	// 初始化 GitHub Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	Client = c.NewClient(tc)

	// 获取 Members 成员列表
	LoadMembers()
	// 获取 Team 成员列表
	LoadMaintainers()
}
