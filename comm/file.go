package comm

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/config"
	"issue-man/global"
	"issue-man/tools"
	"net/http"
)

type File struct {
	PrNumber   int
	CommitFile *github.CommitFile
}

// 处理需同步文件
func (f File) Sync(include config.Include, existIssue, preIssue *github.Issue) {
	// 这里的操作指的是文件的操作
	// 至于 issue 是否存在，调用何种方法，需要额外判断
	const (
		ADD    = "added"
		MODIFY = "modified"
		RENAME = "renamed"
		REMOVE = "removed"
	)
	switch *f.CommitFile.Status {
	// 更新 issue，不存在则创建 issue
	case ADD, MODIFY:
		// 更新 issue
		if existIssue != nil {
			f.update(existIssue)
		} else {
			// 创建 issue
			f.create(include)
		}
	// 重命名/移动文件
	case RENAME:
		f.rename(include, existIssue, preIssue)
	// 移除文件
	case REMOVE:
		f.remove(existIssue)
	default:
		global.Sugar.Warnw("unknown status",
			"file", f,
			"status", *f.CommitFile.Status)
	}
}

// 更新 issue，并 comment 如果 issue 不存在，则创建
func (f File) create(include config.Include) {
	issue := tools.Generate.NewIssue(include, *f.CommitFile.Filename)
	_, resp, err := global.Client.Issues.Create(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		issue,
	)
	if err != nil {
		global.Sugar.Errorw("sync create issues",
			"step", "create",
			"title", issue.Title,
			"body", issue.Body,
			"err", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("init issues",
			"step", "create",
			"title", issue.Title,
			"body", issue.Body,
			"status code", resp.StatusCode,
			"resp body", string(body))
		return
	}
}

// 更新 issue，并 comment
func (f File) update(existIssue *github.Issue) (*github.Issue, error) {
	// 更新
	updatedIssue, err := f.edit(
		tools.Generate.UpdateIssue(false, f.CommitFile.GetFilename(), *existIssue),
		existIssue.GetNumber(),
		"update",
	)
	if err != nil {
		return nil, err
	}

	// 仅 comment 特定状态（label）下的 issue
	if f.commentVerify(existIssue) {
		// comment
		err = f.comment(existIssue)
		if err != nil {
			return nil, err
		}
	}
	return updatedIssue, nil
}

func (f File) commentVerify(issue *github.Issue) bool {
	if issue == nil || issue.Labels == nil {
		return false
	}

	return false
}

// 取 issue 的 number 和 assignees 调用 api 进行 comment
// comment 内容为相关文件改动的提示
func (f File) comment(issue *github.Issue) error {
	// comment
	body := ""
	bf := bytes.Buffer{}
	bf.WriteString("maintainer: ")
	for _, v := range issue.Assignees {
		bf.WriteString(fmt.Sprintf("@%s ", v.GetLogin()))
	}
	// TODO 抽取配置
	bf.WriteString(fmt.Sprintf("\nstatus: %s", f.CommitFile.GetStatus()))
	bf.WriteString(fmt.Sprintf("\npr: https://github.com/istio/istio.io/pull/%d", f.PrNumber))
	bf.WriteString(fmt.Sprintf("\ndiff: https://github.com/istio/istio.io/pull/%d/files#diff-%s",
		f.PrNumber, fmt.Sprintf("%x", md5.Sum([]byte(f.CommitFile.GetFilename())))))
	body = bf.String()

	comment := &github.IssueComment{}
	comment.Body = &body
	_, resp, err := global.Client.Issues.CreateComment(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		issue.GetNumber(),
		comment)

	if err != nil {
		global.Sugar.Errorw("sync issue comment",
			"step", "call api",
			"status", "fail",
			"file", f,
			"err", err.Error())
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("CheckCount",
			"step", "parse response",
			"status", "fail",
			"statusCode", resp.StatusCode,
			"body", string(body))
		return err
	}
	return nil
}

// 删除 issue 中的文件，更新后 issue 内无文件的情况，添加特殊 label 标识，maintainer 手动处理
func (f File) remove(issue *github.Issue) {
	if issue == nil {
		global.Sugar.Warnw("remove exist file issue",
			"status", "has no match issue",
			"file", f)
		return
	}
	updatedIssue, err := f.edit(
		tools.Generate.UpdateIssue(true, f.CommitFile.GetPreviousFilename(), *issue),
		issue.GetNumber(),
		"remove",
	)

	if err != nil {
		return
	}

	// comment
	err = f.comment(updatedIssue)
	if err != nil {
		return
	}
}

// 对于 renamed 文件，需要：
// 1. 更新/创建 新的 issue
// 2. 在旧的 issue 中移除对应的文件
func (f File) rename(include config.Include, existIssue, preIssue *github.Issue) {
	// 更新 issue
	if existIssue != nil {
		// preIssue 为空，则仅更新 existIssue
		// 这种极端情况很难出现
		if preIssue == nil {
			f.update(existIssue)
			global.Sugar.Warnw("renamed file issue",
				"status", "has no match previous issue",
				"file", f)
			return
		}
		// existIssue 和 preIssue 是同一个 issue
		if existIssue.GetNumber() == preIssue.GetNumber() {
			// 由于是同一个 issue，可以一次性完成更新，移除
			updatedIssue, err := f.edit(
				tools.Generate.UpdateIssueRequest(true, f.CommitFile.GetPreviousFilename(), tools.Generate.UpdateIssue(false, *f.CommitFile.Filename, *existIssue)),
				existIssue.GetNumber(),
				"renamed",
			)
			if err != nil {
				return
			}
			// comment
			f.comment(updatedIssue)
		} else {
			// 由于 existIssue 和 preIssue 不是同一个 issue
			// 需要分别完成更新、移除
			f.update(existIssue)
			f.remove(preIssue)
		}
	} else {
		// existIssue == nil，
		// 此时，创建 issue，并在 preIssue 中移除旧文件名
		f.create(include)
		// 尝试移除
		if preIssue != nil {
			f.remove(preIssue)
		}
	}
}

func (f File) edit(issue *github.IssueRequest, number int, option string) (*github.Issue, error) {
	updatedIssue, resp, err := global.Client.Issues.Edit(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		number,
		issue,
	)
	if err != nil {
		global.Sugar.Errorw("init issues",
			"step", "update",
			"id", number,
			"title", issue.Title,
			"body", issue.Body,
			"err", err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("edit issues",
			"step", option,
			"id", number,
			"title", issue.Title,
			"body", issue.Body,
			"status code", resp.StatusCode,
			"resp body", string(body))
		return nil, err
	}
	return updatedIssue, nil
}
