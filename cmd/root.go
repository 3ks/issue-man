package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"issue-man/config"
	"issue-man/global"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	//  如果不想在命令行中指定 Client Person Token，也可以选择在环境变量内指定。
	// 环境变量名为 GITHUB_TOKEN
	IssueManToken = "GITHUB_TOKEN"
)

var (
	// token ，支持通过命令行参数或者环境变量指定
	// 不支持写在配置文件内
	token string

	// 指定配置文件路径，默认为 ./config.yaml
	c string

	// 配置文件，同时也包含了 issue 处理流程
	conf *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "issue-man",
	Short: "issue 生命周期的机器人",
	Long:  `issue-man 一个用于管理 Client Issue 生命周期的机器人。`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 通用的加载配置文件、初始化 log 组件函数
func loadAndInit() config.Config {
	// 如果 token 为空，则尝试从环境变量读取 token
	if token == "" {
		token = os.Getenv(IssueManToken)
		// 没有 token 不能启动
		if token == "" {
			fmt.Printf("please input token with argument --token\n")
			os.Exit(0)
		}
	}
	// 如果配置文件为空，则自动尝试读取 ./ 目录下的 config，
	if c == "" {
		c = "./config.yaml"
	}

	// 读取配置文件
	conf = &config.Config{}
	conf.IssueComments = make([]*config.IssueComment, 0)
	conf.Jobs = make([]*config.Job, 0)

	// 读取配置文件
	data, err := afero.ReadFile(afero.NewOsFs(), c)
	if err != nil {
		fmt.Printf("unable to load config file, %v\n", err)
		os.Exit(1)
	}
	bf := bytes.NewBuffer(data)
	// 拆分配置
	cfgs := strings.Split(bf.String(), "---")

	// 遍历读取
	for i := 0; i < len(cfgs); i++ {
		b := &config.Base{}
		err := yaml.Unmarshal([]byte(cfgs[i]), b)
		if err != nil {
			panic(err.Error())
		}
		switch b.Kind {
		// Repository 的配置
		case "Repository":
			tmp := &config.Repository{}
			err = yaml.Unmarshal([]byte(cfgs[i]), tmp)
			if err != nil {
				panic(err.Error())
			}
			conf.Repository = tmp
		// IssueCreate 的配置
		case "IssueCreate":
			tmp := &config.IssueCreate{}
			err = yaml.Unmarshal([]byte(cfgs[i]), tmp)
			if err != nil {
				panic(err.Error())
			}
			conf.IssueCreate = tmp
		// IssueComment 的配置
		case "IssueComment":
			tmp := &config.IssueComment{}
			err = yaml.Unmarshal([]byte(cfgs[i]), tmp)
			if err != nil {
				panic(err.Error())
			}
			conf.IssueComments = append(conf.IssueComments, tmp)
		// Job 的配置
		case "Job":
			tmp := &config.Job{}
			err = yaml.Unmarshal([]byte(cfgs[i]), tmp)
			if err != nil {
				panic(err.Error())
			}
			conf.Jobs = append(conf.Jobs, tmp)
		// 不支持类型的配置
		default:
			fmt.Printf("Unsupport Type: %s\n", b.Kind)
		}
	}

	// 初始化 Client Client，初始化一些全局变量，其中一些信息需调用 Client API
	global.Init(token, conf)

	// 返回配置对象
	return *conf
}
