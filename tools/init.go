// tools 包封装了一些通用的方法，tools 包只提供方法，没有函数
// 调用方法类似于 tools.<option>.method()
// option 是在 init 文件初始化的一系列可导出变量
// 其作用相当于将方法做了分类，仅此而已，无它。
// 每一类具体实现了哪些方法，可以在对应的 .go 文件中查看。
package tools

var (
	Convert  convertFunctions
	Verify   verifyFunctions
	Get      getFunctions
	Parse    parseFunctions
	Generate generateFunctions
)

type (
	// 封装了一些解析相关的方法
	parseFunctions byte

	// 封装了一些验证相关的方法
	verifyFunctions byte

	// 封装了一些转换，添加，移除相关的方法
	convertFunctions byte

	// 封装了一些获取相关的方法
	getFunctions byte

	// 封装了一些生成内容的方法
	// generate 和 get 方法的区别是， get 用于生成内建数据类型
	// generate 生成的内容一般是自定义 struct 类型
	generateFunctions byte
)
