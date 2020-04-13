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
		needDoSomething, err := judgeState(ss[0], ss[1], job.InstructName, *v.Number, job.Labels, job.In)
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

func judgeState(owner, repository, instructName string, issueNumber int, labels []string, in int64) (needDoSomething bool, err error) {
	createdAt, err := getLabelCreateAt(owner, repository, issueNumber, labels)
	if err != nil {
		return false, err
	}
	//if time.Now().Sub(*createdAt).Milliseconds() > ((int64(time.Hour) * 24 * in) / 1000 / 1000) {
	if time.Now().Sub(createdAt.AddDate(0, 0, int(in))) > 0 {
		// 再查询是否有延迟的指令
		if ins, ok := config.Instructions[instructName]; !ok {
			// 没有相关的延迟指令
			// 可能需要做点什么
			fmt.Printf("need do something\n")
			return true, nil
		} else {
			// 有相关的延迟指令，则判断指令的时间
			// 获取最后执行该指令的时间
			commentAt, err := LastDelayAt(owner, repository, issueNumber, instructName)
			// 没有执行过相关指令，且现在的时间已经大于预期时间，则返回 true
			if err != nil {
				fmt.Printf("can not found comment, err: %v", err.Error())
				fmt.Printf("need do something\n")
				return true, nil
			}
			// 执行过相关指令，且找到了，且期望的时间已经过了，则需要做点什么
			if time.Now().Sub(commentAt.AddDate(0, 0, int(ins.Delay))) > 0 {
				fmt.Printf("need do something\n")
				return true, nil
			} else {
				// 执行过相关指令，且找到了，且期望的时间还没过，则不需要做什么
				fmt.Printf("nothing to do\n")
				return false, nil
			}
		}
	}
	// 时辰未到
	fmt.Printf("nothing to do\n")
	return false, nil
}

// 查询 issue comment，看是否有 `/delay-reset` 指令。
// 如果有，则根据最终一次 comment 的时间，加上 delay 的天数，得出一个重置日期
// 计算当前时间是否大于期望的重置七日
func LastDelayAt(owner, repository string, issueNumber int, instructName string) (commentAt *time.Time, err error) {
	comments, resp, err := client.Get().Issues.ListComments(context.TODO(), owner, repository, issueNumber, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get list comments failed. status code: %v\n", resp.StatusCode)
	}

	for i := len(comments) - 1; i >= 0; i-- {
		if strings.Contains(*comments[i].Body, instructName) {
			return comments[i].CreatedAt, nil
		}
	}
	return nil, fmt.Errorf("not found instruct: %v in issue: %v\n", instructName, issueNumber)
}

func getLabelCreateAt(owner, repository string, issueNumber int, labels []string) (createdAt *time.Time, err error) {
	es, resp, err := client.Get().Issues.ListIssueEvents(context.TODO(), owner, repository, issueNumber, nil)
	if err != nil {
		fmt.Printf("get list events by issue fail. err: %v\n", err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
		return nil, fmt.Errorf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
	}

	for i := len(es) - 1; i >= 0; i-- {
		// 打 label 操作
		if *es[i].Event == Labeled {
			// 找到对应操作的 event
			// todo 暂时视作仅一个 label
			if *es[i].Label.Name == labels[0] {
				return es[i].CreatedAt, nil
			}
		}
	}
	return nil, fmt.Errorf("not found this state")
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
