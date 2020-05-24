package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"issue-man/config"
	"os"
	"path"
	"strings"
)

const (
	IssueManToken = "ISSUE_MAN_TOKEN"
)

var (
	// token ，仅可用个命令行输入
	token string

	// 指定配置文件路径，默认为 ./config.yaml
	c string

	// 配置文件，同时也包含了 issue 处理流程
	conf *config.Config2
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "GitHub Person Token.")
	rootCmd.PersistentFlags().StringVarP(&c, "config", "c", "", "指定配置文件路径")
}

func start() {
	// 如果配置文件为空，则自动尝试读取 ./ 目录下的 config，
	// 后缀可以是 yaml, toml, json 等
	if c == "" {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		c = "./config.yaml"
	} else {
		// 如果指定了配置文件，则读取配置文件
		viper.SetConfigName(strings.TrimSuffix(path.Base(c), path.Ext(c)))
		viper.AddConfigPath(path.Dir(c))
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("read in config fail. config: %v,  err: %v", strings.TrimSuffix(path.Base(c), path.Ext(c)), err))
	}

	// 如果 token 为空，则尝试从环境变量读取 token
	if token == "" {
		token = os.Getenv(IssueManToken)
		// 没有 token 不能启动
		if token == "" {
			fmt.Printf("please input token with argument --token\n")
			os.Exit(0)
		}
	}

	// 读取配置
	initConf()

	// 重定向标准版输出和标准错误
	//openStdFile()

	// 初始化服务相关的东西
	//server.Start(token)
}

// 重定向标准输出和标准错误
func openStdFile() {
	fd, err := os.OpenFile(path.Join(conf.LogDir, conf.StdOutFile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		_ = os.Mkdir(conf.LogDir, os.ModeDir)
		fd, err = os.OpenFile(path.Join(conf.LogDir, conf.StdOutFile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("open stdout file fail. err: %v", err.Error())
			os.Exit(1)
		}
	}
	os.Stdout = fd
	os.Stderr = fd
}

func initConf() {
	// 读取配置文件
	conf = &config.Config2{}
	//err := viper.Unmarshal(&conf)
	//if err != nil {
	//	fmt.Printf("unable to decode into struct, %v\n", err)
	//	os.Exit(1)
	//}
	data, err := afero.ReadFile(afero.NewOsFs(), c)
	if err != nil {
		fmt.Printf("unable to load config file, %v\n", err)
		os.Exit(1)
	}
	bf := bytes.NewBuffer(data)
	cfgs := strings.Split(bf.String(), "---")
	fmt.Println(len(cfgs))
	for i := 0; i < len(cfgs); i++ {

	}

	// todo 对一些默认值进行处理
	if conf.LogDir == "" {
		conf.LogDir = "./log"
	}
	if conf.LogFile == "" {
		conf.LogFile = "issue-man.log"
	}
	if conf.StdOutFile == "" {
		conf.StdOutFile = "issue-man.std.log"
	}
	if conf.Port == "" {
		conf.Port = ":8080"
	}
	fmt.Printf("load config: %#v\n", conf)
	viper.Set("config", conf)
}
