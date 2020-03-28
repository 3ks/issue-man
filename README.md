# Issue Man

一个用于管理 GitHub Issue 生命周期的机器人。

# 配置化的工作流程

Issue Man 的工作流程完全由配置文件决定，您可以根据自己的需要，编写自己特有的流程。

# Issue Man 执行过程

- 解析 webhook 数据，通过 [go-playground/webhooks](https://github.com/go-playground/webhooks) 实现。
- 拼装数据，根据 GitHub API 要求，以及自身需要拼装数据。
- 发送请求，通过 [go-github](https://github.com/google/go-github) 实现。
