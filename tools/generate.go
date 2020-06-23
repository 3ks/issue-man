package tools

import (
	"bytes"
	"fmt"
	"issue-man/global"
	"path"
	"sort"
	"strings"
)

// parseTitleFromPath 解析路径，生成 title
// 传入的路径总是这样的：content/en/faq/setup/k8s-migrating.md，预期 title 为： faq/setup
// 对于文件名为：_index 开头的文件，预期 title 总是为： Architecture
// 不会出现返回 nil 的情况，最差情况下返回值为 ""
func (g generateFunctions) Title(filePath string) *string {
	title := ""
	// TODO 1 以目录、文件为单位（配置化）进行划分
	// TODO 2 抽取 _index？
	if strings.ReplaceAll(path.Base(filePath), path.Ext(filePath), "") == "_index" {
		title = "Architecture"
		return &title
	}
	filePath = strings.Replace(filePath, "content/en/", "", 1)
	t := strings.Split(filePath, "/")
	title = strings.Join(t[:len(t)-1], "/")
	return &title
}

// parseURLFormPath
// 根据 PATH 生成站点的  HTTPS URL
func (g generateFunctions) URL(filePath string) (source, translate string) {
	// 去除两端路径
	url := strings.Split(strings.Replace(filePath, global.Conf.Repository.Spec.Source.RemovePrefix, "", 1), "/")
	tmp := path.Join(url[:len(url)-1]...)

	sourceSite := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(global.Conf.Repository.Spec.Source.Site, "/"), "https://"), "http://")
	translateSite := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(global.Conf.Repository.Spec.Translate.Site, "/"), "https://"), "http://")

	return fmt.Sprintf("https://%s", path.Join(sourceSite, tmp)), fmt.Sprintf("https://%s", path.Join(translateSite, tmp))
}

// genBody 根据文件名和旧的 body，生成新的 body
func (g generateFunctions) Body(remove bool, file, oldBody string) (body *string, length int) {
	t := ""
	body = &t
	oldBody = strings.ReplaceAll(oldBody, "\r\n", "\n")

	// map 存储去重
	files := make(map[string]string)
	files[file] = file
	lines := strings.Split(oldBody, "\n")
	for _, line := range lines {
		if strings.Contains(line, "content/en") { // TODO
			// 去掉旧文件的 https://xxx.com 前缀，后面会重新生成
			tmp := strings.Split(line, "content/en")
			if len(tmp) != 2 {
				continue
			}
			line = fmt.Sprintf("content/en%s", tmp[1])
			files[line] = line
		}
	}
	// 用于移除某个文件的情况
	if remove {
		delete(files, file)
	}

	length = len(files)
	fs := make([]string, len(files))
	// map 转 slice 以便排序
	count := 0
	for k := range files {
		fs[count] = k
		count++
	}
	// 排序
	sort.Slice(fs, func(i, j int) bool {
		return fs[i] < fs[j]
	})

	source, translate := g.URL(file)

	// 构造 body
	bf := bytes.Buffer{}
	// _index 类文件无统一页面
	if strings.Contains(file, "_index") {
		bf.WriteString(fmt.Sprintf("## Source\n\n#### Files\n\n"))
	} else {
		bf.WriteString(fmt.Sprintf("## Source\n\n#### URL\n\n%s\n\n#### Files\n\n", source))
	}
	for _, v := range fs {
		if v == "" {
			continue
		}
		bf.WriteString(fmt.Sprintf("- https://github.com/%s/%s/tree/master/%s\n\n",
			global.Conf.Repository.Spec.Source.Owner,
			global.Conf.Repository.Spec.Source.Repository,
			v))
	}

	bf.WriteString("\n")

	// _index 类文件无统一页面
	if strings.Contains(file, "_index") {
		bf.WriteString(fmt.Sprintf("## Translate\n\n#### Files\n\n"))
	} else {
		bf.WriteString(fmt.Sprintf("## Translate\n\n#### URL\n\n%s\n\n#### Files\n\n", translate))
	}
	for _, v := range fs {
		if v == "" {
			continue
		}
		bf.WriteString(fmt.Sprintf("- https://github.com/%s/%s/tree/master/%s\n\n",
			global.Conf.Repository.Spec.Translate.Owner,
			global.Conf.Repository.Spec.Translate.Repository,
			strings.ReplaceAll(v, "content/en", "content/zh"))) // TODO
	}
	t = bf.String()
	return
}
