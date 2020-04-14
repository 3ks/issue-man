package operation

import (
	"context"
	"fmt"
	"issue-man/client"
	"issue-man/config"
	"net/http"
	"strings"
	"time"

	gg "github.com/google/go-github/v30/github"
)

// Job
// 获取 issue
// 获取 events
// 判断状态、时间。已经处于该状态的则忽略（todo 或其它策略）
// 拼装 Flow 格式的对象，调用相关方法，更新 issue 或 card
func Job(fullName string, job config.Job) {
	fmt.Printf("====================\ndo job: %s\n", job.Name)
	defer fmt.Printf("====================\nfinished job: %s\n", job.Name)
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
			// 缺少任何一个 label，则需要处理，不 continue
			if _, ok := lm[label]; !ok {
				skip = false
			}
		}
		if skip {
			continue
		}

		// 告警
		if job.Name == "warn" {
			// 处于 waiting-for-pr 状态 in 天后，计算预期重置时间，并提示，并添加 stale。
			// 判断进入该状态的时长
			createdAt, err := getLabelCreateAt(ss[0], ss[1], *v.Number, job.Labels)
			if err != nil {
				fmt.Printf("get label create at failed. label: %v, err: %v\n", job.Labels, err.Error())
				continue
			}
			// 预期打上 stale 的时间
			exceptedTime := createdAt.AddDate(0, 0, int(job.In))

			fmt.Printf("wating-for-pr created at: %v, excepted labe stale at: %v, will not label: %v\n", createdAt.String(), exceptedTime.String(), time.Now().Sub(exceptedTime) < 0)
			// 暂不添加 stale 标签，下次一定
			// 当前时间 < 应该打上 stale 的时间，则下次一定
			if time.Now().Sub(exceptedTime) < 0 {
				continue
			}

			// 由于此时还没有 stale 标签，所以，重置时间是当前时间 + reset 的时间
			info, flow := assemblyData(*v, job, fullName)
			assign := ""
			if len(v.Assignees) > 0 {
				assign = *v.Assignees[0].Login
			}
			reset := config.Jobs["reset"]
			hc := Comment{}
			hc.Login = assign
			hc.ResetDate = time.Now().AddDate(0, 0, int(reset.In)).In(time.Local).Format("2006-01-02")

			// 清楚 comment，下面手动调用 comment
			flow.SuccessFeedback = ""
			// 添加标签，后续根据标签和延长指令来计算重置时间
			IssueEdit(info, flow)
			fmt.Printf("warn comment issue number: %v, body: %v\n", info.IssueNumber, hc.HandComment(job.Feedback))
			// comment 提示
			IssueComment(info, hc.HandComment(job.Feedback))
		} else {

			// reset 重置/告警

			// 计算最终重置时间
			// 如果符合重置条件，则重置，并 continue
			// 否则，再判断是否符合提醒条件，如果符合，则提醒
			// 获取 events，判断状态持续时间。
			ins := config.Instructions[job.InstructName]
			resetDate, err := getResetDate(ss[0], ss[1], *v.Number, job.Labels, int(job.In), int(ins.Delay), ins.Name)
			if err != nil {
				fmt.Printf("get and judge issue event failed. err: %v\n", err.Error())
				continue
			}
			resetDate = resetDate.In(time.Local)

			// 当前时间大于重置时间，则应该重置
			if time.Now().Sub(resetDate) > 0 {
				// 需要做点啥，满足条件的，执行操作
				// 拼装一个 info 和 flow，直接调用函数
				info, flow := assemblyData(*v, job, fullName)
				fmt.Printf("assembly data. info: %#v, flow: %#v\n", info, flow)

				// 发送 Update Issue 请求（如果有的话）
				IssueEdit(info, flow)

				// 发送 Move Card 请求（如果有的话）
				CardMove(info, flow)
				continue
			}

			assign := ""
			if len(v.Assignees) > 0 {
				assign = *v.Assignees[0].Login
			}
			// （可能需要）提醒
			remind(ss[0], ss[1], *v.Number, assign, job.InstructName, job.WarnFeedback, resetDate)
		}
	}
}

// 根据剩余时间计算是否需要提醒
// 每天最多提醒一次
func remind(owner, repository string, issueNumber int, login, instructName, warn string, resetDate time.Time) {
	// 计算今天是否需要提醒
	hour := resetDate.In(time.Local).Sub(time.Now()).Hours()
	day := 1
	need := false
	fmt.Printf("last %v hour.\n", hour)
	for i := 0; i < 10; i++ {
		left := (day - 1) * 24
		right := day * 24
		if float64(left) < hour && hour < float64(right) {
			fmt.Printf("last %v hour. bewteen %v and %v\n", hour, left, right)
			need = true
			break
		}
		day *= 2
	}
	if !need {
		return
	}

	// 根据 `/delay-reset` 及 comment 时间判断今天是否已经提醒过了
	remindAt, err := LastRemindAt(owner, repository, issueNumber, instructName)
	if err != nil {
		fmt.Printf("it's ok. get last remind time fail. err: %v\n", err.Error())
	}
	// 存在，且在今天提示过，则不再重复提示
	if remindAt != nil {
		now := time.Now()
		// 今天已提醒过
		// todo 测试时可注释掉
		if now.Year() == remindAt.Year() && now.Month() == remindAt.Month() && now.Day() == remindAt.Day() {
			fmt.Printf("already remind today. date: %v\n", remindAt.String())
			return
		}
	}
	// 其它情况（未提示过，或者提示过，但不是在今天），则提示
	info := Info{
		Owner:       owner,
		Repository:  repository,
		IssueNumber: issueNumber,
	}
	hc := Comment{}
	hc.ResetDate = resetDate.In(time.Local).Format("2006-01-02")
	hc.Login = login
	fmt.Printf("reset comment issue number: %v, body: %v\n", info.IssueNumber, hc.HandComment(warn))
	IssueComment(info, hc.HandComment(warn))
}

// 为 job 拼装一个 info 和 flow
func assemblyData(issue gg.Issue, job config.Job, fullName string) (info Info, flow config.Flow) {
	// info
	ss := strings.SplitN(fullName, "/", -1)
	info.Owner = ss[0]
	info.Repository = ss[1]

	info.Login = *issue.User.Login
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

// 查询 issue comment，看是否有 `/delay-reset` 指令。
// 如果有，则根据最终一次 comment 的时间，加上 delay 的天数，得出一个重置日期
// 计算当前时间是否大于期望的重置七日
func LastDelayAt(owner, repository string, issueNumber, delay int, instructName string) (commentAt time.Time, err error) {
	comments, resp, err := client.Get().Issues.ListComments(context.TODO(), owner, repository, issueNumber, nil)
	if err != nil {
		return commentAt, err
	}
	if resp.StatusCode != http.StatusOK {
		return commentAt, fmt.Errorf("get list comments failed. status code: %v\n", resp.StatusCode)
	}

	for i := len(comments) - 1; i >= 0; i-- {
		// 以指令开头
		if strings.HasPrefix(*comments[i].Body, instructName) {
			return comments[i].CreatedAt.AddDate(0, 0, delay), nil
		}
	}
	return commentAt, fmt.Errorf("last delay at. not found instruct: %v in issue: %v\n", instructName, issueNumber)
}

// 获取最后一次提醒的时间
func LastRemindAt(owner, repository string, issueNumber int, instructName string) (commentAt *time.Time, err error) {
	comments, resp, err := client.Get().Issues.ListComments(context.TODO(), owner, repository, issueNumber, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get list comments failed. status code: %v\n", resp.StatusCode)
	}

	for i := len(comments) - 1; i >= 0; i-- {
		// todo 怎么识别提示
		if strings.Contains(*comments[i].Body, "will be reset") {
			tmp := comments[i].CreatedAt.In(time.Local)
			return &tmp, nil
			// 不是单纯的指令，还包含提示字符，视为一次提醒
			//if len(*comments[i].Body) > len(instructName) {
			//	return comments[i].CreatedAt, nil
			//}
		}
	}
	return nil, fmt.Errorf("last remind at. not found instruct: %v in issue: %v\n", instructName, issueNumber)
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
				tmp := es[i].CreatedAt.In(time.Local)
				return &tmp, nil
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
