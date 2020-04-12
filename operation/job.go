package operation

import (
	"context"
	"fmt"
	gg "github.com/google/go-github/v30/github"
	"issue-man/client"
	"issue-man/config"
	"net/http"
	"strings"
	"time"
)

// Job
// 获取 issue
// 获取 events
// 判断状态、时间。已经处于该状态的则忽略（todo 或其它策略）
// 拼装 Flow 格式的对象，调用相关方法，更新 issue 或 card
func Job(fullName string, job config.Job) {
	fmt.Printf("do job: %#v\n", job)
	ss := strings.SplitN(fullName, "/", -1)
	if len(ss) != 2 {
		fmt.Printf("unknown repository: %s\n", fullName)
		return
	}

	// 获取满足条件的 issue
	issues, err := getIssues(ss[0], ss[1], job.Labels)
	if err != nil {
		fmt.Printf("get issues list with label failed. label: %v, err: %v\n", job.Labels, err.Error())
		return
	}
	fmt.Printf("get issues with label: %#v. list: %#v\n", job.Labels, issues)

	for _, v := range issues {
		lm := make(map[string]bool)
		skip := true

		// 当前已有的 label
		for _, label := range v.Labels {
			lm[*label.Name] = true
		}
		// 如果已经有目标 label，则跳过？
		for _, label := range job.TargetLabels {
			// 缺少任何一个 label，则需要处理
			if _, ok := lm[label]; !ok {
				skip = false
			}
		}
		if skip {
			continue
		}

		// 获取 events，判断状态持续时间。（以打上 label 的时间为准）
		needDoSomething, err := judgeState(ss[0], ss[1], *v.Number, job.Labels, job.In)
		if err != nil {
			fmt.Printf("get and judge issue event failed. err: %v\n", err.Error())
			continue
		}

		// 不需要再做啥了，下一个
		if !needDoSomething {
			continue
		}

		// 需要做点啥，满足条件的，执行操作
		// 拼装一个 info 和 flow，直接调用函数
		info, flow := assemblyData(*v, job, fullName)
		fmt.Printf("assembly data. info: %#v, flow: %#v\n", info, flow)

		// 发送 Update Issue 请求（如果有的话）
		IssueEdit(info, flow)

		// 发送 Move Card 请求（如果有的话）
		CardMove(info, flow)
	}
}

// 为 job 拼装一个 info 和 flow
func assemblyData(issue gg.Issue, job config.Job, fullName string) (info Info, flow config.Flow) {
	// info
	ss := strings.SplitN(fullName, "/", -1)
	info.Owner = ss[0]
	info.Repository = ss[1]

	info.Title = *issue.Title
	info.Body = *issue.Body
	info.Milestone = *issue.Milestone.Number
	info.Assignees = make([]string, len(issue.Assignees))
	info.Labels = make([]string, len(issue.Labels))
	for i := 0; i < len(issue.Assignees) || i < len(issue.Labels); i++ {
		if i < len(info.Assignees) {
			info.Assignees[i] = *issue.Assignees[i].Login
		}
		if i < len(info.Labels) {
			info.Labels[i] = *issue.Labels[i].Name
		}
	}
	info.IssueURL = *issue.URL
	info.IssueNumber = *issue.Number
	info.State = *issue.State

	// assign 策略
	switch job.AssigneesPolicy {
	case "@remove_all":
		info.Assignees = make([]string, 0)
	case "@keep_all":
	default:
	}
	// todo Login 暂时用 assignees[0] 填充
	if len(info.Assignees) != 0 {
		info.Login = info.Assignees[0]
	}

	// flow
	flow.Close = false
	flow.RemoveLabel = job.RemoveLabels
	flow.TargetLabel = job.TargetLabels
	flow.SuccessFeedback = job.Feedback
	flow.CurrentColumnID = job.CurrentColumnID
	flow.TargetColumnID = job.TargetColumnID
	return
}

const (
	Labeled = "labeled"
)

func judgeState(owner, repository string, issueNumber int, labels []string, in int64) (needDoSomething bool, err error) {
	es, resp, err := client.Get().Issues.ListIssueEvents(context.TODO(), owner, repository, issueNumber, nil)
	if err != nil {
		fmt.Printf("get list events by issue fail. err: %v\n", err.Error())
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
		return false, fmt.Errorf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
	}

	for i := len(es) - 1; i >= 0; i-- {
		// 打 label 操作
		if *es[i].Event == Labeled {
			// 找到对应操作的 event
			// todo 暂时视作仅一个 label
			if *es[i].Label.Name == labels[0] {
				// 判断持续时间
				if time.Now().Sub(*es[i].CreatedAt).Milliseconds() > ((int64(time.Hour) * 24 * in) / 1000 / 1000) {
					// 需要做点什么
					fmt.Printf("need do something\n")
					return true, nil
				} else {
					// 时辰未到
					fmt.Printf("nothing to do\n")
					return false, nil
				}
			}
		}
	}
	return false, nil
}

func getIssues(owner, repository string, labels []string) (issues []*gg.Issue, err error) {
	// 根据 label 筛选
	req := &gg.IssueListByRepoOptions{}
	req.Labels = labels

	is, resp, err := client.Get().Issues.ListByRepo(context.TODO(), owner, repository, req)
	if err != nil {
		fmt.Printf("get list issue by repo fail. err: %v\n", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
		return
	}
	return is, err
}
