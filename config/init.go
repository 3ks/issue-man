package config

import (
	"github.com/spf13/viper"
)

// Maintainers
var Maintainers map[string]bool

// 指令及其对应的 Flow
// 每个指令对应一个 Flow
// 支持哪些指令，取决于配置文件内容
// Flow 的工作流程，取决于配置文件内容
// Flow 的行为及处理逻辑，取决于配置文件
var Instructions map[string]Flow

// Jobs
var Jobs map[string]Job

func Init() {
	conf, ok := viper.Get("config").(*Config2)
	if !ok {
		panic("viper get config fail")
	}

	Maintainers = make(map[string]bool)
	Jobs = make(map[string]Job)
	Instructions = make(map[string]Flow)

	// maintain map
	for _, v := range conf.Maintainers {
		Maintainers[v] = true
	}

	// flow map
	for _, v := range conf.Flows {
		Instructions[v.Name] = v
	}

	// job map
	for _, v := range conf.Jobs {
		if v.In < 0 {
			continue
		}
		Jobs[v.Name] = v
	}
}
