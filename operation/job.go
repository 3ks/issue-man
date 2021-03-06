package operation

/*
import (
	"context"
	"fmt"
	"issue-man/config"
	"issue-man/global"
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
func Job(fullName string, Sync config.Job) {
	fmt.Printf("====================\ndo Sync: %s\n", Sync.Name)
	defer fmt.Printf("====================\nfinished Sync: %s\n", Sync.Name)
	ss := strings.SplitN(fullName, "/", -1)
	if len(ss) != 2 {
		fmt.Printf("unknown repository: %s\n", fullName)
		return
	}

	// 获取满足条件的 issue
	issues, err := GetAllIssue(ss[0], ss[1], Sync.Labels)
	if err != nil {
		fmt.Printf("get issues list with label failed. label: %v, err: %v\n", Sync.Labels, err.Error())
		return
	}
	fmt.Printf("get issues with label: %#v. list: %#v\n", Sync.Labels, issues)

	for _, v := range issues {
		lm := make(map[string]bool)
		skip := true

		// 当前已有的 label
		for _, label := range v.Labels {
			lm[*label.Name] = true
		}
		// 如果已经有目标 label，则跳过？
		for _, label := range Sync.TargetLabels {
			// 缺少任何一个 label，则需要处理，不 continue
			if _, ok := lm[label]; !ok {
				skip = false
			}
		}
		if skip {
			continue
		}

		// 告警
		if Sync.Name == "warn" {
			// 处于 waiting-for-pr 状态 in 天后，计算预期重置时间，并提示，并添加 stale。
			// 判断进入该状态的时长
			createdAt, err := getLabelCreateAt(ss[0], ss[1], *v.Number, Sync.Labels)
			if err != nil {
				fmt.Printf("get label create at failed. label: %v, err: %v\n", Sync.Labels, err.Error())
				continue
			}
			// 预期打上 stale 的时间
			exceptedTime := createdAt.AddDate(0, 0, int(Sync.In))

			fmt.Printf("wating-for-pr created at: %v, excepted labe stale at: %v, will not label: %v\n", createdAt.Strings(), exceptedTime.Strings(), time.Now().Sub(exceptedTime) < 0)
			// 暂不添加 stale 标签，下次一定
			// 当前时间 < 应该打上 stale 的时间，则下次一定
			if time.Now().Sub(exceptedTime) < 0 {
				continue
			}

			// 由于此时还没有 stale 标签，所以，重置时间是当前时间 + reset 的时间
			info, flow := assemblyData(*v, Sync, fullName)
			assign := ""
			if len(v.Assignees) > 0 {
				assign = *v.Assignees[0].Login
			}
			reset := global.Jobs["reset"]
			hc := Comment{}
			hc.Login = assign
			hc.ResetDate = time.Now().AddDate(0, 0, int(reset.In)).In(time.Local).Format("2006-01-02")

			// 清楚 comment，下面手动调用 comment
			flow.SuccessFeedback = ""
			// 添加标签，后续根据标签和延长指令来计算重置时间
			issueEdit(info, flow)
			fmt.Printf("warn comment issue number: %v, body: %v\n", info.IssueNumber, hc.HandComment(Sync.Feedback))
			// comment 提示
			IssueComment(info, hc.HandComment(Sync.Feedback))
		} else {

			// reset 重置/告警

			// 计算最终重置时间
			// 如果符合重置条件，则重置，并 continue
			// 否则，再判断是否符合提醒条件，如果符合，则提醒
			// 获取 events，判断状态持续时间。
			ins := global.Instructions[Sync.InstructName]
			resetDate, err := getResetDate(ss[0], ss[1], *v.Number, Sync.Labels, int(Sync.In), int(ins.Delay), ins.Name)
			if err != nil {
				fmt.Printf("get and judge issue event failed. err: %v\n", err.Error())
				continue
			}
			resetDate = resetDate.In(time.Local)

			// 当前时间大于重置时间，则应该重置
			if time.Now().Sub(resetDate) > 0 {
				// 需要做点啥，满足条件的，执行操作
				// 拼装一个 info 和 flow，直接调用函数
				info, flow := assemblyData(*v, Sync, fullName)
				fmt.Printf("assembly data. info: %#v, flow: %#v\n", info, flow)

				// 发送 Update Issue 请求（如果有的话）
				issueEdit(info, flow)

				// 发送 Move Card 请求（如果有的话）
				CardMove(info, flow)
				continue
			}

			assign := ""
			if len(v.Assignees) > 0 {
				assign = *v.Assignees[0].Login
			}
			// （可能需要）提醒
			remind(ss[0], ss[1], *v.Number, assign, Sync.InstructName, Sync.WarnFeedback, resetDate)
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
			fmt.Printf("already remind today. date: %v\n", remindAt.Strings())
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

// 为 Sync 拼装一个 info 和 flow
func assemblyData(issue gg.Issue, Sync config.Job, fullName string) (info Info, flow config.Flow) {
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
	switch Sync.AssigneesPolicy {
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
	flow.RemoveLabel = Sync.RemoveLabels
	flow.TargetLabel = Sync.TargetLabels
	flow.SuccessFeedback = Sync.Feedback
	flow.CurrentColumnID = Sync.CurrentColumnID
	flow.TargetColumnID = Sync.TargetColumnID
	return
}

const (
	Labeled = "labeled"
)

// 获取某个 issue 的 comment 列表
// 查询 issue comment，看是否有 `/delay-reset` 指令。
// 如果有，则根据最终一次 comment 的时间，加上 delay 的天数，得出并返回一个重置日期
func LastDelayAt(owner, repository string, issueNumber, delay int, instructName string) (commentAt time.Time, err error) {
	comments := make([]*gg.IssueComment, 0)
	req := &gg.IssueListCommentsOptions{}
	req.PerPage = 100

	for i := 1; i <= 100; i++ {
		cs, resp, err := global.Get().Issues.ListComments(context.TODO(), owner, repository, issueNumber, req)
		if err != nil {
			return commentAt, err
		}
		if resp.StatusCode != http.StatusOK {
			return commentAt, fmt.Errorf("get list comments failed. status code: %v\n", resp.StatusCode)
		}
		// 无记录，终止
		if len(cs) == 0 {
			break
		}
		comments = append(comments, cs...)
		// 应该没有下一页了
		if len(cs) < req.PerPage {
			break
		}
	}

	for i := len(comments) - 1; i >= 0; i-- {
		// 以指令开头
		if strings.HasPrefix(*comments[i].Body, instructName) {
			return comments[i].CreatedAt.AddDate(0, 0, delay), nil
		}
	}
	return commentAt, fmt.Errorf("last delay at. not found instruct: %v in issue: %v\n", instructName, issueNumber)
}

// 获取某个 issue 的 comment 列表
// 并判断最后一次提醒的时间
func LastRemindAt(owner, repository string, issueNumber int, instructName string) (commentAt *time.Time, err error) {
	comments := make([]*gg.IssueComment, 0)
	req := &gg.IssueListCommentsOptions{}
	req.PerPage = 100

	for i := 1; i <= 100; i++ {
		cs, resp, err := global.Get().Issues.ListComments(context.TODO(), owner, repository, issueNumber, req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("get list comments failed. status code: %v\n", resp.StatusCode)
		}
		// 无记录，终止
		if len(cs) == 0 {
			break
		}
		comments = append(comments, cs...)
		// 应该没有下一页了
		if len(cs) < req.PerPage {
			break
		}
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

// 获取某个 issue 的 event 列表
// 并判断某些 label 的创建时间
func getLabelCreateAt(owner, repository string, issueNumber int, labels []string) (createdAt *time.Time, err error) {
	// 会先判断 issue 目前是否含有期望的 label
	ls, _, err := global.Get().Issues.Get(context.TODO(), owner, repository, issueNumber)
	if err != nil {
		fmt.Printf("get issue by issue_number fail. err: %v\n", err.Error())
		return nil, err
	}
	lsm := make(map[string]bool)
	for _, v := range ls.Labels {
		lsm[*v.Name] = true
	}
	// 没有期望的 label，则不再查询 events
	if _, ok := lsm[labels[0]]; !ok {
		return nil, fmt.Errorf("labels not exist now,nothing to do")
	}

	events := make([]*gg.IssueEvent, 0)
	req := &gg.ListOptions{}
	req.PerPage = 100

	for i := 1; i <= 100; i++ {
		req.Page = i
		es, resp, err := global.Get().Issues.ListIssueEvents(context.TODO(), owner, repository, issueNumber, req)
		if err != nil {
			fmt.Printf("get list events by issue fail. err: %v\n", err.Error())
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
			return nil, fmt.Errorf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
		}
		// 无记录，终止
		if len(es) == 0 {
			break
		}
		events = append(events, es...)
		// 应该没有下一页了
		if len(es) < req.PerPage {
			break
		}
	}

	for i := len(events) - 1; i >= 0; i-- {
		// 打 label 操作
		if *events[i].Event == Labeled {
			// 找到对应操作的 event
			// todo 暂时视作仅一个 label
			if *events[i].Label.Name == labels[0] {
				tmp := events[i].CreatedAt.In(time.Local)
				return &tmp, nil
			}
		}
	}
	return nil, fmt.Errorf("not found this state")
}

// 获取 issue 列表
// 获取全部（10000 条以内）满足条件的 issue
func GetAllIssue(owner, repository string, labels []string) (issues []*gg.Issue, err error) {
	issues = make([]*gg.Issue, 0)

	// 根据 label 筛选
	req := &gg.IssueListByRepoOptions{}
	req.Labels = labels
	req.PerPage = 100

	for i := 1; i < 100; i++ {
		// 页数
		req.Page = i
		is, resp, err := global.Get().Issues.ListByRepo(context.TODO(), owner, repository, req)
		if err != nil {
			fmt.Printf("get list issue by repo fail. err: %v\n", err.Error())
			break
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("list issue by repo maybe fail. status code: %v\n", resp.StatusCode)
			break
		}

		// 无记录，终止
		if len(is) == 0 {
			break
		}
		issues = append(issues, is...)
		// 应该没有下一页了
		if len(is) < req.PerPage {
			break
		}
	}

	return issues, err
}
*/
