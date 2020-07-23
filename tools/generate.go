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
// 如果配置文件指预期 title 为： faq/setup
// 对于文件名为：_index 开头的文件，预期 title 总是为： Architecture
// 不会出现返回 nil 的情况，最差情况下返回值为 ""
func (g generateFunctions) Title(filename string, include config.Include) *string {

	// 如果 include.Title 有值，则不计算 title，直接返回该值
	if include.Title != "" {
		return Get.String(include.Title)
	}

	// 按照目录分类命名
	genTitleByDirectory := func(filename string) *string {
		if !global.Conf.IssueCreate.Spec.SaveTitlePrefix {
			filename = strings.TrimPrefix(filename, global.Conf.IssueCreate.Spec.Prefix)
		}
		return Get.String(path.Dir(filename))
	}

	// 按文件分类命名
	genTitleByFile := func(filename string) *string {
		if !global.Conf.IssueCreate.Spec.SaveTitlePrefix {
			filename = strings.TrimPrefix(filename, global.Conf.IssueCreate.Spec.Prefix)
		}
		return Get.String(filename)
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
	url := strings.Split(strings.Replace(filePath, global.Conf.Repository.Spec.Source.RemovePrefix, "", 1), "/")
	tmp := path.Join(url[:len(url)-1]...)
	sourceSite := strings.TrimPrefix(strings.TrimPrefix(global.Conf.Repository.Spec.Source.Site, "https://"), "http://")
	translateSite := strings.TrimPrefix(strings.TrimPrefix(global.Conf.Repository.Spec.Translate.Site, "https://"), "http://")

	return fmt.Sprintf("https://%s", path.Join(sourceSite, tmp)), fmt.Sprintf("https://%s", path.Join(translateSite, tmp))
}

// genBody 根据文件名和旧的 body，生成新的 body
func (g generateFunctions) Body(remove bool, file, oldBody string) (body *string, length int) {
	t := ""
	body = &t
	oldBody = strings.ReplaceAll(oldBody, "\r\n", "\n")

	// map 存储去重
	files := g.extractFilesFromBody(oldBody)
	// 用于移除某个文件的情况
	files[file] = true
	if remove {
		delete(files, file)
	}
	length = len(files)

	fs := Convert.MapToString(files)
	source, translate := g.URL(file)

	// 如果 URL 数量大于 1，则不构造 URL 和 Commits History 内容

	// 反之，则构造

	// 构造 body
	bf := bytes.Buffer{}
	// _index 类文件无统一页面
	if strings.Contains(file, "_index") {
		bf.WriteString(fmt.Sprintf("## Source\n\n#### Files\n\n"))
	} else {
		// TODO 配置化
		bf.WriteString(fmt.Sprintf("## Source\n\nURL：%s/%s\n\nHistory：%s\n\nFile：",
			source, strings.TrimSuffix(path.Base(file), path.Ext(file)),
			fmt.Sprintf("https://github.com/%s/%s/commits/master/%s\n\n",
				global.Conf.Repository.Spec.Source.Owner,
				global.Conf.Repository.Spec.Source.Repository,
				file,
			),
		))
	}
	for _, v := range *fs {
		if v == "" {
			continue
		}
		bf.WriteString(fmt.Sprintf("https://github.com/%s/%s/tree/master/%s\n\n",
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			v))
	}

	bf.WriteString("\n")

	// _index 类文件无统一页面
	if strings.Contains(file, "_index") {
		bf.WriteString(fmt.Sprintf("## Translate\n\n#### Files\n\n"))
	} else {
		bf.WriteString(fmt.Sprintf("## Translate\n\nURL：%s/%s\n\nHistory：%s\n\nFile：",
			translate, strings.TrimSuffix(path.Base(file), path.Ext(file)),
			fmt.Sprintf("https://github.com/%s/%s/commits/master/%s\n\n",
				global.Conf.Repository.Spec.Translate.Owner,
				global.Conf.Repository.Spec.Translate.Repository,
				file,
			),
		))
	}
	for _, v := range *fs {
		if v == "" {
			continue
		}
		bf.WriteString(fmt.Sprintf("https://github.com/%s/%s/tree/master/%s\n\n",
			global.Conf.Repository.Spec.Translate.Owner,
			global.Conf.Repository.Spec.Translate.Repository,
			strings.ReplaceAll(v, "content/en", "content/zh"))) // TODO
	}
	t = bf.String()
	return
}

// extractFilesFromBody 提取 body 内的文件列表
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
