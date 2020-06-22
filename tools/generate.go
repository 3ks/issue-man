package tools

import (
	"fmt"
	"issue-man/global"
	"path"
	"strings"
)

// parseTitleFromPath 解析路径，生成 title
// 传入的路径总是这样的：content/en/faq/setup/k8s-migrating.md，预期 title 为： faq/setup
// 对于文件名为：_index 开头的文件，预期 title 总是为： Architecture
// 不会出现返回 nil 的情况，最差情况下返回值为 ""
func (p parseFunctions) Title(filePath string) *string {
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
func (p parseFunctions) URL(filePath string) (source, translate string) {
	// 去除两端路径
	url := strings.Split(strings.Replace(filePath, global.Conf.Repository.Spec.Source.RemovePrefix, "", 1), "/")
	tmp := path.Join(url[:len(url)-1]...)

	sourceSite := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(global.Conf.Repository.Spec.Source.Site, "/"), "https://"), "http://")
	translateSite := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(global.Conf.Repository.Spec.Translate.Site, "/"), "https://"), "http://")

	return fmt.Sprintf("https://%s", path.Join(sourceSite, tmp)), fmt.Sprintf("https://%s", path.Join(translateSite, tmp))
}
