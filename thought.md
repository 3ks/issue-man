## 流程

issue-man 的工作流程。

- 解析 webhook 数据，通过 [go-playground/webhooks](https://github.com/go-playground/webhooks) 实现。
- 拼装数据，根据 GitHub API 要求，以及自身需要拼装数据。
- 发送请求，通过 [go-github](https://github.com/google/go-github) 实现。

## 思路

在
