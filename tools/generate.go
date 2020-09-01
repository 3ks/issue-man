package tools

import (
	"bytes"
	"fmt"
	"github.com/google/go-github/v30/github"
	"issue-man/config"
	"issue-man/global"
	"path"
	"strings"
)

// 分类依据
const (
	Directory = "directory"
	File      = "file"
)

// parseTitleFromPath 解析路径，生成 title
// 传入的路径总是这样的：content/en/faq/setup/k8s-migrating.md
// 如果 include.Title != "" 则直接返回 include.Title 的值
func (g generateFunctions) Title(filename string, include config.Include) *string {

	// 如果 include.Title 有值，则不计算 title，直接返回该值
	if include.Title != "" {
		return Get.String(include.Title)
	}

	// 按照目录分类命名
	genTitleByDirectory := func(filename string) *string {
		// 不去除前缀
		if global.Conf.IssueCreate.Spec.SaveTitlePrefix {
			return Get.String(path.Dir(filename))
		}
		return Get.String(path.Dir(strings.TrimPrefix(filename, global.Conf.IssueCreate.Spec.Prefix)))
	}

	// 按文件分类命名
	genTitleByFile := func(filename string) *string {
		// 不去除前缀
		if global.Conf.IssueCreate.Spec.SaveTitlePrefix {
			Get.String(filename)
		}
		return Get.String(strings.TrimPrefix(filename, global.Conf.IssueCreate.Spec.Prefix))
	}

	switch {
	case include.GroupBy == File:
		return genTitleByFile(filename)
	case include.GroupBy == Directory:
		return genTitleByDirectory(filename)
	case global.Conf.IssueCreate.Spec.GroupBy == File:
		return genTitleByFile(filename)
	default:
		return genTitleByDirectory(filename)
	}
}

// parseURLFormPath
// 根据 PATH 生成站点的  HTTPS URL
func (g generateFunctions) URL(filePath string) (source, translate string) {
	// TODO 不同项目 URL 命名规则不同
	// TODO
	// 去除两端路径
	// 根据 RemovePrefix 去除前缀
	url := strings.Split(strings.Replace(filePath, global.Conf.IssueCreate.Spec.Prefix, "", 1), "/")
	// 去掉最后一段
	tmp := path.Join(url[:len(url)-1]...)

	// 去掉 Site 中的 https:// 和 http:// 字符
	sourceSite := strings.TrimPrefix(strings.TrimPrefix(global.Conf.Repository.Spec.Source.Site, "https://"), "http://")
	translateSite := strings.TrimPrefix(strings.TrimPrefix(global.Conf.Repository.Spec.Translate.Site, "https://"), "http://")

	return fmt.Sprintf("https://%s", path.Join(sourceSite, tmp)), fmt.Sprintf("https://%s", path.Join(translateSite, tmp))
}

// genBody 根据文件名和旧的 body，生成新的 body
// 按文件
// URL：site+removeSuffix(removePrefix(file))（根据配置文件决定是否移除）
// HISTORY：owner+repository+commit+file
// FILE：owner+repository+tree+file
//
// 按目录
// URL：site+removeSuffix(removePrefix(file))（根据配置文件决定是否移除）
// HISTORY：owner+repository+commit+dir(file)
// FILES：
// - owner+repository+tree+file
// - owner+repository+tree+file
// - owner+repository+tree+file
func (g generateFunctions) Body(remove bool, file, oldBody string) (body *string, length int) {
	// 构造 body
	bf := bytes.Buffer{}

	// Requirement 必要信息
	require := fmt.Sprintf("### Requirement\n\n翻译人员信息登录：%s\n\n翻译指南：%s\n\n",
		"https://baidu.com",
		"https://baidu.com",
	)
	bf.WriteString(require)

	// 网页 URL，此时的 URL 可能不完整，可能需要拼接文件名（一般会需要除非后缀，比如 .md 等），也可能不需要拼接文件名（比如按目录划分 URL）
	sourceSiteURL, translateSiteURL := g.URL(file)

	// 按文件分隔，此时直接构造 body
	if global.Conf.IssueCreate.Spec.GroupBy == File {
		if remove {
			return nil, 0
		}
		// Source URL
		url := fmt.Sprintf("[envoyproxy.io/docs](%s/%s)", sourceSiteURL, strings.TrimSuffix(path.Base(file), ".rst.txt"))
		// Source FileCommitHistory
		history := fmt.Sprintf("[envoyproxy/envoyproxy.github.io#FileCommitHistory](https://github.com/%s/%s/commits/%s/%s)\n\n",
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			global.Conf.Repository.Spec.Source.Branch,
			file,
		)
		// Source FILE
		filename := fmt.Sprintf("[%s](https://github.com/%s/%s/tree/%s/%s)\n\n",
			file,
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			global.Conf.Repository.Spec.Source.Branch,
			file)

		// Source
		bf.WriteString(fmt.Sprintf("## Source\n\nURL：%s\n\nHistory：%s\n\nFile：%s\n\n", url, history, filename))

		// Translate URL
		url = fmt.Sprintf("[cloudnative.to/envoy/docs](%s/%s)", translateSiteURL, strings.TrimSuffix(path.Base(file), ".rst.txt"))

		// Translate FileCommitHistory
		history = fmt.Sprintf("[cloudnative/envoy#FileCommitHistory](https://github.com/%s/%s/commits/%s/%s)\n\n",
			global.Conf.Repository.Spec.Translate.Owner,
			global.Conf.Repository.Spec.Translate.Repository,
			global.Conf.Repository.Spec.Translate.Branch,
			file,
		)
		// Translate FILE
		filename = fmt.Sprintf("[%s](https://github.com/%s/%s/tree/%s/%s)\n\n",
			file,
			global.Conf.Repository.Spec.Translate.Owner,
			global.Conf.Repository.Spec.Translate.Repository,
			global.Conf.Repository.Spec.Translate.Branch,
			file)

		// Translate
		bf.WriteString(fmt.Sprintf("## Translate\n\nURL：%s\n\nHistory：%s\n\nFile：%s\n\n", url, history, filename))
		return Get.String(bf.String()), 1
	}

	// 按目录
	// URL：site+removeSuffix(removePrefix(file))（根据配置文件决定是否移除）
	// HISTORY：owner+repository+commit+dir(file)
	// FILES：
	// - owner+repository+tree+file
	// - owner+repository+tree+file
	// - owner+repository+tree+file
	// 否则按目录分割，此时需要提取 oldBody 的文件列表，并更新
	// map 存储去重
	files := g.extractFilesFromBody(strings.ReplaceAll(oldBody, "\r\n", "\n"))
	// 用于移除某个文件的情况
	files[file] = true
	if remove {
		delete(files, file)
	}
	length = len(files)

	// 文件列表
	fileSlice := Convert.MapToString(files)

	// Source
	// Source URL，按目录分的 resp，其网页一般不带文件名
	url := fmt.Sprintf("[envoyproxy.io/docs](%s)", sourceSiteURL)

	// Source FileCommitHistory
	history := fmt.Sprintf("[envoyproxy/envoyproxy.github.io#FileCommitHistory](https://github.com/%s/%s/commits/%s/%s\n\n)",
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		global.Conf.Repository.Spec.Source.Branch,
		path.Dir(file), // 目录
	)
	bf.WriteString(fmt.Sprintf("## Source\n\nURL：%s\n\nHistory：%s\n\n", url, history))

	// Source FILES
	bf.WriteString("Files：\n")
	for _, v := range *fileSlice {
		if v == "" {
			continue
		}
		// File URL in GitHub
		bf.WriteString(fmt.Sprintf("- [%s](https://github.com/%s/%s/tree/%s/%s)\n",
			v,
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			global.Conf.Repository.Spec.Source.Branch,
			v))
	}

	// Translate
	// Translate URL，按目录分的 resp，其网页一般不带文件名
	url = fmt.Sprintf("[cloudnative.to/envoy/docs](%s)", sourceSiteURL)

	// Translate FileCommitHistory
	history = fmt.Sprintf("[cloudnative/envoy#FileCommitHistory](https://github.com/%s/%s/commits/%s/%s\n\n)",
		global.Conf.Repository.Spec.Translate.Owner,
		global.Conf.Repository.Spec.Translate.Repository,
		global.Conf.Repository.Spec.Translate.Branch,
		path.Dir(file), // 目录
	)
	bf.WriteString(fmt.Sprintf("\n## Translate\n\nURL：%s\n\nHistory：%s\n\n", url, history))

	// Translate FILES
	bf.WriteString("Files：\n")
	for _, v := range *fileSlice {
		if v == "" {
			continue
		}
		// Translate File URL in GitHub
		bf.WriteString(fmt.Sprintf("- [%s](https://github.com/%s/%s/tree/%s/%s)\n",
			v,
			global.Conf.Repository.Spec.Translate.Owner,
			global.Conf.Repository.Spec.Translate.Repository,
			global.Conf.Repository.Spec.Translate.Branch,
			v))
	}

	return Get.String(bf.String()), len(*fileSlice)
}

// extractFilesFromBody 提取 body 内的文件列表
// 要求每个文件一行，并且
// 包含 global.Conf.IssueCreate.Spec.Prefix 关键字
// 提取的内容为 Prefix 之后的内容
// map 存储去重
func (g generateFunctions) extractFilesFromBody(body string) (files map[string]bool) {
	files = make(map[string]bool)
	lines := strings.Split(body, "\n")
	prefix := global.Conf.IssueCreate.Spec.Prefix
	for _, line := range lines {
		index := strings.Index(line, prefix)
		if index > 0 {
			// 去掉旧文件 prefix 及前面的内容（https://xxx.com/xxx/），后面会重新生成
			line = line[index+len(prefix):]
			files[line] = true
		}
	}
	return
}

// BodyByPRNumberAndSha
// 根据 pr Number 和 sha 生成 issue body
// BodyByPRNumberAndSha() 有一个对应的解析方法 PRNumberFromBody()
func (g generateFunctions) BodyByPRNumberAndSha(number int, sha string) *string {
	body := fmt.Sprintf("https://github.com/%s/%s/pull/%d\n\nhttps://github.com/%s/%s/tree/%s",
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		number,
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		sha,
	)
	return &body
}

// 根据配置初始化一个新 issue 的内容
func (g generateFunctions) NewIssue(include config.Include, filename string) (new *github.IssueRequest) {
	new = &github.IssueRequest{}
	new.Title = Generate.Title(filename, include)
	new.Body, _ = Generate.Body(false, filename, "")

	new.Labels = Convert.SliceAdd(&include.Labels, global.Conf.IssueCreate.Spec.Labels...)
	new.Assignees = Get.Strings(global.Conf.IssueCreate.Spec.Assignees)
	new.Milestone = Get.Int(global.Conf.IssueCreate.Spec.Milestone)
	return
}

// 更新已存在 issue
func (g generateFunctions) UpdateIssue(remove bool, file string, exist github.Issue) (update *github.IssueRequest) {
	length := 0
	update = &github.IssueRequest{}
	update.Title = exist.Title
	update.Body, length = Generate.Body(remove, file, *exist.Body)

	// 对于已存在的 issue
	// label、assignees、milestone 不会变化
	if exist.Milestone != nil {
		update.Milestone = exist.Milestone.Number
	}
	if exist.Labels != nil {
		update.Labels = Convert.Label(exist.Labels)
	}
	if exist.Assignees != nil {
		update.Assignees = Convert.Assignees(exist.Assignees)
	}

	// 如果文件列表为 0，则添加需要检查的 Label
	// 反之则移除
	if length == 0 {
		update.Labels = Convert.SliceAdd(update.Labels, global.Conf.Repository.Spec.Workspace.Detection.DeprecatedLabel...)
	} else {
		update.Labels = Convert.SliceRemove(update.Labels, global.Conf.Repository.Spec.Workspace.Detection.DeprecatedLabel...)
	}
	return
}

// 根据已存在的 issue 和配置，返回更新后的 issue
// 仅更新 body
func (g generateFunctions) UpdateNewIssue(file string, exist *github.IssueRequest) *github.IssueRequest {
	exist.Body, _ = Generate.Body(false, file, exist.GetBody())
	return exist
}

// 更新 issue request 的 body
func (g generateFunctions) UpdateIssueRequest(remove bool, file string, exist *github.IssueRequest) *github.IssueRequest {
	exist.Body, _ = Generate.Body(remove, file, exist.GetBody())
	return exist
}
