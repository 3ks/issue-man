// 更新或评论 issue
package operation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	client http.Client
	header http.Header
)

type Issue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Milestone int      `json:"milestone,omitempty"`
	State     string   `json:"state"`
	Assignees []string `json:"assignees,omitempty"`
	Labels    []string `json:"labels"`
	URL       URLs     `json:"-"`
}

type Commenter struct {
	Login string `json:"login"`
	Body  string `json:"body"`
}

type URLs struct {
	RepositoryURL  string `json:"-"`
	RepositoryName string `json:"-"`

	ID          int    `json:"-"` // IssueID
	IssueURL    string `json:"-"`
	CommentsURL string `json:"-"`
}

// 更新 Issue
// 获取原始数据
// /accept: 评论、修改 label、分配任务
// /pushed: 评论、修改 label
// /merged: 评论、修改label
func UpdateIssue(v Issue) (err error) {

	data, _ := json.Marshal(v)
	req, _ := http.NewRequest(http.MethodPost, v.URL.IssueURL, bytes.NewReader(data))
	req.Header = header
	response, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("issue update id: %v, full: %#v, url: %v, do request err: %v\n", v.URL.ID, v, v.URL.IssueURL, err.Error())
		fmt.Println(err.Error())
		return
	}

	// 修改成功
	if response.StatusCode == http.StatusOK {
		return nil
	}

	// 解析错误
	resp, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		err = fmt.Errorf("issue update id: %v, full: %#v, url: %v, read resp body err: %v\n", v.URL.ID, v, v.URL.IssueURL, err.Error())
		fmt.Println(err.Error())
		return
	}
	err = fmt.Errorf("issue update id: %v, full: %#v, url: %v, get resp err: %s\n", v.URL.ID, v, v.URL.IssueURL, string(resp))
	fmt.Println(err.Error())
	return
}

type Comment struct {
	Body string `json:"body"`
	URL  URLs   `json:"-"`
}

// 评论 issue，并 @ 操作人
func CommentIssue(v Comment) {

	data, _ := json.Marshal(v)
	req, _ := http.NewRequest(http.MethodPost, v.URL.CommentsURL, bytes.NewReader(data))

	req.Header = header
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("issue comment id: %v, full: %#v, url: %v, do request err: %v\n", v.URL.ID, v, v.URL.CommentsURL, err.Error())
		return
	}

	// 创建成功
	if response.StatusCode == http.StatusCreated {
		return
	}
	resp, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("issue comment id: %v, full: %#v, url: %v, read resp body err: %v\n", v.URL.ID, v, v.URL.CommentsURL, err.Error())
		return
	}
	fmt.Printf("issue comment id: %v, full: %#v, url: %v, get resp err: %s\n", v.URL.ID, v, v.URL.CommentsURL, string(resp))
}

// 查询 accept 的 issue 数量
// 处于 open 状态
// assign 给 login
// 处于 waiting-for-pr 状态
func GetIssueCount(url, login string) int {

	url = fmt.Sprintf("%s?state=%s&assignee=%s&labels=%s", url, "open", login, SWaiting)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header = header

	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("query issues: %v, do request err: %v\n", url, err.Error())
		return 0
	}

	resp, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		fmt.Printf("query issues: %v, read resp err: %v\n", url, err.Error())
		return 0
	}

	issues := make([]QueryIssue, 0)
	err = json.Unmarshal(resp, &issues)
	if err != nil {
		fmt.Printf("query issues: %v, unmarshal resp err: %v\n", url, err.Error())
		return 0
	}

	return len(issues)
}
