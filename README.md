# Issue Man

Issue Man 是一个通过配置文件定义并管理 GitHub Issue 生命周期的机器人。

# 配置化的工作流程

Issue Man 的工作流程完全由配置文件决定，您可以根据自己的需要，编写自己特有的流程。Issue Man 在工作时服务的仓库，仓库的管理员，每条指令的名称，对权限的要求，数量限制，涉及的 label，assignees 改动，issue 的关闭，提示文本等。所有的内容都是通过配置文件来定义的。

虽然 Issue Man 很多选项都可以自行配置，但这不代表一定需要配置，很多配置项都是可以缺省不填的，对此，Issue Man 不会做其它额外的动作，只会安装你定义的指令来处理。

这儿有一个简单的流程定义示例：

```yaml
full_repository_name: "gorda/gorda.io"
maintains:
  - "gorda"
flows:
  # /accept 指令
  - name: "/accept"
    permission:
      - "maintainers"
      - "member"
    current_label:
      - "status/spending"
    target_label:
      - "status/waiting-for-pr"
    success_feedback: "Thank you @somebody, this issue had been assigned to you."

  # /pushed 指令
  - "name": "/pushed"
    permission:
      - "maintainers"
      - "self"
    current_label:
      - "status/waiting-for-pr"
    target_label:
      - "status/reviewing"

  # /merged 指令
  - "name": "/merged"
    permission:
      - "maintainers"
      - "self"
    current_label:
      - "status/reviewing"
    target_label:
      - "status/merged"
    close: true
```

其包含了三条 `指令`：`/accept`，`/pushed`，`/merged`。这是一种常见的基于 issue 的工作流程。
据我所知，[ServiceMesher](https://github.com/servicemesher) 和 [k8smeetup](https://github.com/k8smeetup) 都是采用的这种工作流程。

除此之外，你还可以根据自己的需要自定义任何指令及其操作。
在了解 Issue Man 的基本用法后，你可以查看[指令文档](instruction.md)了解完整的 Issue Man 目前支持的配置项。

# Issue Man 执行过程

- 解析 webhook 数据，通过 [go-playground/webhooks](https://github.com/go-playground/webhooks) 实现。
- 拼装数据，根据 GitHub API 要求，以及自身需要拼装数据。
- 发送请求，通过 [go-github](https://github.com/google/go-github) 实现。
