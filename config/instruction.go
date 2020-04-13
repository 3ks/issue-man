package config

// Flow 定义了一个指令的工作流程
// 每个指令对应一个 Flow
// 支持哪些指令，取决于配置文件内容
// Flow 的工作流程，取决于配置文件内容
// Flow 的行为及处理逻辑，取决于配置文件
type Flow struct {

	// 指令名
	Name string `mapstructure:"name"`

	// 权限验证，指定可以执行该操作的人员。
	// 注意：可以执行该操作，不表示该操作一定会完成，这取决于实际流程限制。
	// 可选的参数有：anyone, member, self, maintainers。默认为空，表示不允许执行指令。
	// issue-man 会按照 maintainers、self、member、anyone 的顺序检查，一旦满足，立即返回 true，否则返回 false。
	// anyone 对执行该指令的人无要求，例如：/cc @someone
	// member 要求执行该指令的人是组织的 member，例如：/accept
	// self 要求执行该指令的人在 assigns 列表中，例如：只能 /unassign 已经分配给自己它的 issue。
	// Maintainers 要求执行改操作的成员在 Maintainers 列表内，例如 /unassign @another，则直接将 another 从 assigns 列表移除。
	// 一般地，Maintainers 成员，总是有足够的权限执行各种指令。（即一般都要为指令带上 Maintainers 参数）
	Permission []string `mapstructure:"permission"`

	// 未通过权限验证时的提示信息，如果为空，则表示不提示，默认为空。
	// issue-man 中的文字，可以传入一些占位符，目前支持的有：@somebody
	// 例如：@somebody thank you!
	// issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。
	PermissionFeedback string `mapstructure:"permission_feedback"`

	// 如何处理 comment 提及的人
	// 可选的值有：addition, remove, none
	// 为 addition，则将提及的人添加至 assignees 列表
	// 为 remove，则将提及的人从 assignees 列表中移除
	// 为 none 则不做处理
	// 为空，或其他值，效果与 none 相同，默认为空。
	Mention string `mapstructure:"mention"`

	// 当前阶段 label
	// 如果 issue 当前不包含这些 label，则表示当前 issue 不满足当前条件，终止操作。
	// 如果列表为空，则表示该指令不涉及移除 label
	CurrentLabel []string `mapstructure:"current_label"`

	// 一般进入下一阶段的状态后，会移除这些 label
	RemoveLabel []string `mapstructure:"remove_label"`

	// 目标状态 label
	// 一般进入下一阶段的状态后，会添加这些 label
	// 如果目标 label 不存在，GitHub 是会自动创建的，也就是说，对于新仓库，其实不需要手动创建 label
	// 如果列表为空，则表示该指令不涉及添加 label
	TargetLabel []string `mapstructure:"target_label"`

	// 成功进入目标状态后的文字提示，如果为空，则表示不提示，默认为空
	// issue-man 中的文字，可以传入一些占位符，目前支持的有：@somebody
	// 例如：@somebody thank you!
	// issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。
	SuccessFeedback string `mapstructure:"success_feedback"`

	// 未能成功进入目标状态时的文字提示，如果为空，则表示不提示，默认为空
	// issue-man 中的文字，可以传入一些占位符，目前支持的有：@somebody
	// 例如：@somebody thank you!
	// issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。
	FailFeedback string `mapstructure:"fail_feedback"`

	// 目标状态的数量限制，0 表示无限制，默认为 0
	// 一般只有表示“处理中”的这种状态的 issue，才需要进行数量的限制。
	// 大不多的阶段都是不需要做限制的，例如“已关闭”，“已结束”等。
	TargetLimit int `mapstructure:"target_limit"`

	// 当触发数量限制时的文字提示，如果为空，则表示不提示，默认为空
	// issue-man 中的文字，可以传入一些占位符，目前支持的有：@somebody @count
	// 例如：@somebody sorry can't accept more than @count issue!
	// issue-man 会自动将 @somebody 替换为 comment 该指令的那个人。
	// issue-man 会自动将 @count 替换为 TargetLimit。
	LimitFeedback string `mapstructure:"limit_feedback"`

	// 指定该指令是否会关闭 issue，一般只有关闭 issue 的指令才会将该值置为 true
	// 默认为 false，即不关闭 issue
	Close bool `mapstructure:"close"`

	// issue 与 project 的关联操作，如果 issue 与 project 需要联动，则需要配置下面的项。
	// 这一部分需要提供一些 ID，实际上，我们在 web 界面看到的数字是 Number，而不是 ID
	// 但是，你可以通过调用 issue-man 提供的接口，获取相关信息，然后根据自己的情况编写配置文件。
	// 为什么需要这些 ID？因为 issue 与 project 的关联性并不强，issue-man 需要手动实现这一部分的逻辑，所以需要提供 ID
	// 根据 GitHub 的结构，在每个 Project 中，可以有多个 column，多个 card。而 card 随着状态的变化，可以属于不同的 column。
	// 同时 card 的 content_url 字段，指向一个 URL 地址。
	// 一般来说，为 issue 创建的 card，该 card 的 content_url 会指向该 issue 的地址（API地址）。
	// issue-man 就是通过 content_url 和 issue url 来判断两者是否存在关联。
	// todo 目前需要每次 get card list，然后遍历判断，后续考虑接受 project 的 webhook，动态维护一份数据。

	// ProjectID
	// 对于将 card 从一个 column 移动到另一个 column 这个操作，不需要用到 ProjectID。
	// ProjectID uint

	// 当前 ColumnID
	CurrentColumnID int64 `mapstructure:"current_column_id"`

	// 目标 ColumnID
	TargetColumnID int64 `mapstructure:"target_column_id"`

	// 目标位置，即选择移动至目标 Column 的顶部还是底部。
	// 支持 top 和 bottom，默认 top。
	// 实际上 GitHub 还支持精确的移动到某个 card 后面，但是这需要提供 card_id，对于用户来说，这太难了。
	TargetPosition string `mapstructure:"target_position"`

	// 移动 card 后的文字提示，为空则不提示，默认为空
	ColumnFeedback string `mapstructure:"column_feedback"`

	// 对应的 job 名
	// 用于找到 job map 中对应的任务，然后获取其重置时间，最终根据 delay 计算过期时间
	JobName string `mapstructure:"job_name"`

	// 推迟天数
	Delay int64 `mapstructure:"delay"`
}
