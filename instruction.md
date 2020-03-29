# 配置项

在 Issue Man 中，配置包含两类，`指令配置` 和 `其它配置`。

对于一个工作流程来说，其核心是对指令进行配置。

除了指令的配置外，Issue Man [其它方面的配置项](#其它方面的配置项)中的 `full_repository_name` 是必填的，表示所服务的仓库的名称：

```yaml
full_repository_name: "gorda/gorda.io"
flows:
...
...
...
```

# 指令目前支持的配置项

在 Issue Man 中，每条指令表示一系列操作，指令均由 GitHub issue Web 界面的用户发出。

Issue Man 在收到 GitHub 用户的指令后，会解析指令，并根据不同的指令，会做出不同的操作。

每条指令的动作完全取决于 Issue Man 所读取到的配置文件，而配置文件，则由你来定义。

### 指令名

键名：name

类型：string

必填：true

```yaml
flows:
  - name: "/accept"
```

> 说明：

指令名，每条指令必须以 `/` 开头

### 权限验证

键名：permission

类型：[]string

必填：true

示例：

```yaml
flows:
  - name: "/accept"
    permission:
      - "maintainers"
      - "self"
```

> 说明：

权限验证，指定可以执行该操作的人员。注意：可以执行该操作，不表示该操作一定会完成，这取决于实际流程限制。

可选的参数有：`anyone`, `member`, `self`, `maintainers`。默认为空，表示不允许任何人执行该指令，也就是说，如果你想让指令工作，则必须配置该项。

issue-man 会按照 maintainers、self、member、anyone 的顺序检查，一旦满足，立即返回 true，否则返回 false。

maintainers 要求执行改操作的成员在 maintainers 列表内，例如 /unassign @another，则直接将 another 从 assigns 列表移除。
一般地，maintainers 成员，总是有足够的权限执行各种指令。（即一般都要为指令带上 maintainers 参数）

self 要求执行该指令的人在 assigns 列表中，例如：只能 /unassign 已经分配给自己它的 issue。

anyone 对执行该指令的人无要求，例如：/cc @someone

member 要求执行该指令的人是组织的 member，例如：/accept

### 权限验证反馈

键名：permission_feedback

类型：string

必填：false

```yaml
flows:
  - name: "/accept"
    permission:
      - "maintainers"
    permission_feedback: "@somebody 抱歉，你没有足够的权限！"
```

> 说明：

未通过权限验证时的提示信息，如果为空，则表示不提示，默认为空。

`permission_feedback` 中的文字，可以传入一些占位符，目前支持的有：@somebody

issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。

### 提及的人

键名：mention

类型：string

必填：false

```yaml
flows:
  - name: "/accept"
    mention: "addition"
```

> 说明：

如何处理 comment 提及的人，可选的值有：`addition`, `remove`, `none`

addition：将提及的人添加至 assignees 列表

remove：将提及的人从 assignees 列表中移除

none：不做处理

为空，或其他值，效果与 none 相同，默认为空。

常见的用法如：`/assign`，`/unassign`，`/cc`

### 当前阶段拥有的 label

键名：current_label

类型：[]string

必填：false

```yaml
flows:
  - name: "/accept"
    current_label:
      - "status/spending"
```

> 说明：

当前阶段拥有的 label，一般进入目标阶段的状态后，会移除这些 label

对于想要在目标阶段保留的 label，你只需将相关的 label 添加到下一阶段的 label 列表即可。

`如果当前 issue 不包含这些 label`，则表示当前 issue 不符合条件，Issue Man 会终止操作。

如果列表为空，则表示该指令不涉及移除 label 和状态检查。

### 目标阶段拥有的 label

键名：target_label

类型：[]string

必填：false

```yaml
flows:
  - name: "/accept"
    current_label:
      - "status/spending"
    target_label:
      - "status/waiting-for-pr"
```

> 说明：

目标阶段 label 列表，一般进入目标阶段的状态后，会添加这些 label

对于上一阶段想要保留的 label，你可以将其添加到 `target_label` 列表

如果 label 列表为空，则表示该指令不涉及添加 label

如果目标 label 不存在，GitHub 是会自动创建的，也就是说，对于新仓库，其实不需要手动创建 label

### 切换状态成功的反馈

键名：success_feedback

类型：string

必填：false

```yaml
flows:
  - name: "/accept"
    current_label:
      - "status/spending"
    target_label:
      - "status/waiting-for-pr"
    success_feedback: "@somebody 干得漂亮！"
```

> 说明：

成功进入目标状态后的文字提示，如果为空，则表示不提示，默认为空。

`success_feedback` 中的文字，可以传入一些占位符，目前支持的有：@somebody

issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。

### 切换状态失败的反馈

键名：fail_feedback

类型：string

必填：false

```yaml
flows:
  - name: "/accept"
    current_label:
      - "status/spending"
    target_label:
      - "status/waiting-for-pr"
    success_feedback: "@somebody 干得漂亮！"
    fail_feedback: "@somebody 抱歉，不可以哦"
```

> 说明：

未能成功进入目标状态后的文字提示，如果为空，则表示不提示，默认为空。

`fail_feedback` 中的文字，可以传入一些占位符，目前支持的有：@somebody

issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。

这里的失败一般指的是 `状态不满足` 导致的失败。 

### 数量限制

键名：target_limit

类型：int

必填：false

```yaml
flows:
  - name: "/accept"
    target_label:
      - "status/waiting-for-pr"
    limit: 10
```

> 说明：

目标状态的数量限制，0 表示无限制，默认为 0

`limit` 需要与 `target_label` 结合使用，表示限制包含某种 label 的 issue 的数量。

一般只有表示 `处理中` 这种含义的 issue，才需要进行数量的限制。

大多数阶段的 issue 是不需要做限制的，例如`已关闭`，`已结束`等。

### 触发限制的反馈

键名：target_limit

类型：int

必填：false

```yaml
flows:
  - name: "/accept"
    target_label:
      - "status/waiting-for-pr"
    limit: 10
    limit_feedback: "@somebody 不可以超过 @count 个哦"
```

> 说明：

当触发数量限制时的文字提示，如果为空，则表示不提示，默认为空

`target_limit` 中的文字，可以传入一些占位符，目前支持的有：@somebody @count

issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。

issue-man 会自动将 @count 替换为 TargetLimit。


### 关闭 issue

键名：close

类型：bool

必填：false

```yaml
flows:
  - name: "/pushed"
    target_label:
      - "status/merged"
    close: true
```

> 说明：

指定该指令是否会关闭 issue，一般只有关闭 issue 的指令才会将该值置为 true

默认为 false，即不关闭 issue

# Project 相关的配置项

这一部分的配置项其实依然属于 `指令配置`，但是这一部分的配置可能用得很少。

所以我将其写在单独的一节，有需要的同学可以看看。

### issue 与 project

issue 与 project 的关联操作，可以实现在 issue 的状态（label）发送变化时，project 中对应的 card 也同步变化（所属的 column）

如果 issue 与 project 需要联动，则需要配置下面的项。

这一部分需要提供一些 `ID`，实际上，我们在 web 界面看到的数字是 Number，而不是 ID，在一个 Project 中，这些 ID 一般是不会有太大变化的

但是，你可以通过调用 issue-man 提供的接口（稍等一下），获取相关 ID，然后根据自己的情况编写配置文件。

为什么需要这些 ID？因为 issue 与 project 的关联性并不强，issue-man 需要手动实现这一部分的逻辑，所以需要提供 ID。

根据 GitHub 的结构，在每个 Project 中，可以有多个 column，多个 card。而 card 随着状态的变化，card 可以属于不同的 column。同时 card 的 content_url 字段，指向一个 URL 地址。

一般来说，为 issue 创建的 card，该 card 的 content_url 会指向该 issue 的地址（API地址）。

issue-man 就是通过 content_url 和 issue url 来判断两者是否存在关联。存在关联时，则操作相关的 card。

### 移动 card

键名：current_column_id 和 target_position

类型：int64

必填：false

```yaml
flows:
  - name: "/accept"
    target_label:
      - "status/waiting-for-pr"
    current_column_id: 9527
    target_column_id: 10086
```

> 说明：

会尝试将当前 issue 对应的 card 从 `current_column_id` 移动至 `target_column_id`。

这两者必须同时使用，否则 Issue Man 不知道你想从哪里移动到哪里。

### 移动至 target_column_id 后的位置

键名：target_position

类型：string

必填：false

```yaml
flows:
  - name: "/accept"
    target_label:
      - "status/waiting-for-pr"
    current_column_id: 9527
    target_column_id: 10086
    target_position: "top"
```

> 说明：

将当前 issue 对应的 card 从 `current_column_id` 移动至 `target_column_id` 时。

card 存放目标位置，即选择移动至 `target_column_id` 的顶部还是底部。支持 `top` 和 `bottom`，默认 `top`。

实际上 GitHub 还支持精确的移动到某个 `card` 后面，但是这需要提供 card_id，对于用户来说，这太难了。

### 移动 card 后的反馈

键名：column_feedback

类型：string

必填：false

```yaml
flows:
  - name: "/accept"
    target_label:
      - "status/waiting-for-pr"
    current_column_id: 9527
    target_column_id: 10086
    column_feedback: "@somebody 该 issue 已从 TODO 移动至 In progress"
```

> 说明：

`column_feedback` 中的文字，可以传入一些占位符，目前支持的有：@somebody 
 
issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。

# 其它方面的配置项

这部分是介绍 Issue Man 中的 `非指令` 配置项，主要是仓库名，服务监听端口，日志目录相关的配置项：

键名                  |  类型   |  必填    | 示例            |  说明       
---------------------|---------|---------|----------------|------------------|
full_repository_name |  string | true    | gorda/gorda.io | 所服务的仓库完整名称
log_dir              |  string | false   | ./log          | 日志目录，默认未 `./log`。主要注意的是，日志文件和标准输出文件都会存放在该目录下，不需要再指定目录名
log_file             |  string | false   | hook.log       | 日志文件，默认为 `issue-man.log`
std_out_file         |  string | false   | hook.std.log   | 标准输出文件，默认为不重定向标准输出
