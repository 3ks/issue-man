package tools

import (
	"context"
	"fmt"
	"issue-man/global"
	"net/http"
)

// GetAllMatchFile
// 基于 tree 获取 resource 中所有符合条件的文件
func (t treeFunctions) GetAllMatchFile(sha string) (files map[string]string, err error) {
	c := *global.Conf
	global.Sugar.Debugw("load upstream files",
		"step", "start")
	ts, resp, err := global.Client.Git.GetTree(context.TODO(),
		c.Repository.Spec.Source.Owner,
		c.Repository.Spec.Source.Repository,
		sha,
		true,
	)
	if err != nil {
		global.Sugar.Errorw("load upstream files",
			"call api", "Get tree",
			"err", err.Error(),
		)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		global.Sugar.Errorw("load upstream files list",
			"call api", "unexpect status code",
			"status", resp.Status,
			"status code", resp.StatusCode,
			"response", resp.Body,
		)
		return nil, fmt.Errorf("%v\n", resp.Body)
	}
	files = make(map[string]string)
	for _, v := range ts.Entries {
		// 仅处理支持的文件类型
		if v.GetType() == "blob" && c.IssueCreate.SupportType(v.GetPath()) {
			files[v.GetPath()] = v.GetPath()
			continue
		}
	}
	//global.Sugar.Debugw("Get files",
	//	"data", files)
	return files, nil
}
