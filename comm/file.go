package comm

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/google/go-github/v30/github"
	"issue-man/config"
	"issue-man/global"
	"issue-man/tools"
)

type File struct {
	PrNumber       int
	MergedAt       string
	MergeCommitSHA string
	CommitFile     *github.CommitFile
}

// 处理需同步文件
func (f File) Sync(include config.Include, existIssue, preIssue *github.Issue) {
	// 这里的操作指的是文件的操作，取值来自于 GitHub
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

// 创建 issue，无 comment
func (f File) create(include config.Include) {
	// 创建通用 issue，按照 create 相关配置初始化、分级
	_, _ = tools.Issue.Create(tools.Generate.NewIssue(include, *f.CommitFile.Filename))
	// 无需 comment
}

// 更新 issue，并 comment
func (f File) update(existIssue *github.Issue) (*github.Issue, error) {
	// 更新
	issue := tools.Generate.UpdateIssue(false, f.CommitFile.GetFilename(), *existIssue)
	// 对于有 assigner 的 issue，添加和移除一些 label
	// 反之，不改动 issue label
	if len(existIssue.Assignees) > 0 {
		issue.Labels = tools.Convert.SliceAdd(issue.Labels, global.Conf.Repository.Spec.Workspace.Detection.AddLabel...)
		issue.Labels = tools.Convert.SliceRemove(issue.Labels, global.Conf.Repository.Spec.Workspace.Detection.RemoveLabel...)
	}

	updatedIssue, err := tools.Issue.EditByIssueRequest(existIssue.GetNumber(), issue)
	if err != nil {
		return nil, err
	}

	// comment
	_ = f.comment(updatedIssue)

	return updatedIssue, nil
}

func (f File) commentVerify(issue *github.Issue) bool {
	if issue == nil || issue.Labels == nil {
		return false
	}
	return tools.Verify.HasAnyLabel(*tools.Convert.Label(issue.Labels), global.Conf.Repository.Spec.Workspace.Detection.NeedLabel...) && len(issue.Assignees) > 0
}

// 取 issue 的 number 和 assignees 调用 api 进行 comment
// comment 内容为相关文件改动的提示
func (f File) comment(issue *github.Issue) error {
	// 对于不满足要求的 issue，不进行 comment
	if !f.commentVerify(issue) {
		return nil
	}
	bf := bytes.Buffer{}
	bf.WriteString(fmt.Sprintf("Pull Request: https://github.com/%s/%s/pull/%d",
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		f.PrNumber))

	bf.WriteString(fmt.Sprintf("\n\nDiff: https://github.com/%s/%s/pull/%d/files#diff-%s",
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		f.PrNumber, fmt.Sprintf("%x", md5.Sum([]byte(f.CommitFile.GetFilename())))))

	bf.WriteString(fmt.Sprintf("\n\nCommit SHA: [%s](https://github.com/%s/%s/blob/%s/%s)",
		f.MergeCommitSHA,
		global.Conf.Repository.Spec.Source.Owner,
		global.Conf.Repository.Spec.Source.Repository,
		f.MergeCommitSHA,
		f.CommitFile.GetFilename(),
	))

	bf.WriteString(fmt.Sprintf("\n\nMerged At: %s", f.MergedAt))

	bf.WriteString(fmt.Sprintf("\n\nFilename: %s", f.CommitFile.GetFilename()))

	if f.CommitFile.GetPreviousFilename() != "" {
		bf.WriteString(fmt.Sprintf("\n\nPrevious Filename: %s", f.CommitFile.GetPreviousFilename()))
	}

	bf.WriteString(fmt.Sprintf("\n\nStatus: %s", f.CommitFile.GetStatus()))

	bf.WriteString("\n\nAssignees: ")
	for _, v := range issue.Assignees {
		bf.WriteString(fmt.Sprintf("@%s ", v.GetLogin()))
	}

	tools.Issue.Comment(issue.GetNumber(), bf.String())
	return nil
}

// 删除 issue 中的文件
func (f File) remove(issue *github.Issue) {
	if issue == nil {
		global.Sugar.Warnw("remove exist file issue",
			"status", "has no match issue",
			"file", f)
		return
	}
	updatedIssue, err := tools.Issue.EditByIssueRequest(issue.GetNumber(), tools.Generate.UpdateIssue(true, f.CommitFile.GetPreviousFilename(), *issue))
	if err != nil {
		return
	}

	// comment
	_ = f.comment(updatedIssue)
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
				"filename", f.CommitFile.GetFilename(),
				"previous filename", f.CommitFile.GetPreviousFilename(),
			)
			return
		}
		// existIssue 和 preIssue 是同一个 issue
		if existIssue.GetNumber() == preIssue.GetNumber() {
			// 由于是同一个 issue，可以一次性完成更新，移除
			updatedIssue, err := tools.Issue.EditByIssueRequest(existIssue.GetNumber(), tools.Generate.UpdateIssueRequest(true, f.CommitFile.GetPreviousFilename(), tools.Generate.UpdateIssue(false, *f.CommitFile.Filename, *existIssue)))
			if err != nil {
				return
			}
			// comment
			_ = f.comment(updatedIssue)
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
