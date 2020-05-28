// comm.go 包含了各个子命令通用的一些遍历
// 如：配置、日志等
package global

import (
	"go.uber.org/zap"
	"issue-man/config"
)

// 各种全局对象
var (
	// 配置对象
	Conf *config.Config

	// 日志对象
	Sugar *zap.SugaredLogger

	// 任务仓库 maintainer 列表
	// 判断某个用户是否为 maintainer 的推荐方法是直接调用 IsMaintainer() 函数，
	// 例如：result:=global.IsMaintainer("gorda")，然后判断返回值即可。
	Maintainers = make(map[string]bool)

	// 组织成员
	// TODO 每次验证不通过，都更新一次 map？以防止成员修改用户名后验证不通过的隐患。
	Member = make(map[string]bool)

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

func Init(conf *config.Config) {
	Conf = conf

	// 从配置文件读取 maintainer 列表
	for _, v := range Conf.Repository.Spec.Maintainers {
		Maintainers[v] = true
	}

	// 从配置文件读取指令列表
	for _, v := range Conf.IssueComments {
		Instructions[v.Metadata.Name] = v
	}

	// 从配置文件读取 Job 列表
	for _, v := range Conf.Jobs {
		if v.Spec.Rules.In == nil || *v.Spec.Rules.In < 0 {
			continue
		}
		Jobs[v.Metadata.Name] = v
	}

	// 生产环境
	if *Conf.Repository.Spec.LogLevel == "pro" {
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err.Error())
		}
		Sugar = logger.Sugar()
	} else {
		// 开发环境
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err.Error())
		}
		Sugar = logger.Sugar()
	}

	Sugar.Infow("finish load config",
		"config", Conf)
}
