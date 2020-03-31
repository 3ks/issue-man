package config

import (
	"github.com/spf13/viper"
)

// 指令及其对应的 Flow
// 每个指令对应一个 Flow
// 支持哪些指令，取决于配置文件内容
// Flow 的工作流程，取决于配置文件内容
// Flow 的行为及处理逻辑，取决于配置文件
var Instructions map[string]Flow

// Maintainers
var Maintainers map[string]bool

func Init() {
	conf, ok := viper.Get("config").(*Config)
	if !ok {
		panic("viper get config fail")
	}

	Instructions = make(map[string]Flow)
	Maintainers = make(map[string]bool)

	// flow map
	for _, v := range conf.Flows {
		Instructions[v.Name] = v
	}

	// maintain map
	for _, v := range conf.Maintainers {
		Maintainers[v] = true
	}
}
