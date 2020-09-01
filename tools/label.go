package tools

import (
	"context"
	"fmt"
	"github.com/google/go-github/v30/github"
	"io/ioutil"
	"issue-man/global"
	"math/rand"
	"net/http"
	"time"
)

// 获取 workspace 的全部 label
func (l labelFunctions) GetAllLabels() (labels []*github.Label, err error) {
	opt := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	labels = make([]*github.Label, 0)

	for {
		tmp, resp, err := global.Client.Issues.ListLabels(
			context.TODO(),
			global.Conf.Repository.Spec.Workspace.Owner,
			global.Conf.Repository.Spec.Workspace.Repository,
			opt,
		)
		if err != nil {
			fmt.Printf("get_label_fail err: %v\n", err.Error())
			global.Sugar.Errorw("label get",
				"step", "call api",
				"status", "fail",
				"err", err.Error())
			return nil, err
		}
		_ = resp.Body.Close()
		labels = append(labels, tmp...)
		if len(tmp) >= 100 {
			opt.Page++
			continue
		}
		break
	}
	return
}

func (l labelFunctions) CreateLabels(label, description string) (err error) {
	// 如果 body 为空，则不做任何操作
	if label == "" {
		return nil
	}
	c := &github.Label{
		Name:  Get.String(label),
		Color: Get.String(getRandomColor()),
	}
	if description != "" {
		c.Description = Get.String(description)
	}

	_, resp, err := global.Client.Issues.CreateLabel(
		context.TODO(),
		global.Conf.Repository.Spec.Workspace.Owner,
		global.Conf.Repository.Spec.Workspace.Repository,
		c,
	)

	if err != nil {
		fmt.Printf("create_label_fail err: %v\n", err.Error())
		global.Sugar.Errorw("label create",
			"step", "call api",
			"status", "fail",
			"label", label,
			"description", description,
			"err", err.Error())
		return
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		global.Sugar.Errorw("label create",
			"step", "check response",
			"label", label,
			"description", description,
			"status code", resp.StatusCode,
			"body", string(body))
		return fmt.Errorf("create label fail. unexcepted status code: %d\n", resp.StatusCode)
	}
	return nil
}

var colors = []string{"FFB6C1", "FFC0CB", "DC143C", "FFF0F5", "DB7093", "FF69B4", "FF1493", "C71585", "DA70D6", "DDA0DD", "EE82EE", "FF00FF", "FF00FF", "8B008B", "800080", "BA55D3", "9400D3", "9932CC", "4B0082", "8A2BE2", "9370DB", "7B68EE", "6A5ACD", "483D8B", "E6E6FA", "F8F8FF", "0000FF", "0000CD", "191970", "00008B", "000080", "4169E1", "6495ED", "B0C4DE", "778899", "708090", "1E90FF", "F0F8FF", "4682B4", "87CEFA", "87CEEB", "00BFFF", "ADD8E6", "B0E0E6", "5F9EA0", "F0FFFF", "E1FFFF", "AFEEEE", "00FFFF", "00FFFF", "00CED1", "2F4F4F", "008B8B", "008080", "48D1CC", "20B2AA", "40E0D0", "7FFFAA", "00FA9A", "F5FFFA", "00FF7F", "3CB371", "2E8B57", "F0FFF0", "90EE90", "98FB98", "8FBC8F", "32CD32", "00FF00", "228B22", "008000", "006400", "7FFF00", "7CFC00", "ADFF2F", "556B2F", "6B8E23", "FAFAD2", "FFFFF0", "FFFFE0", "FFFF00", "808000", "BDB76B", "FFFACD", "EEE8AA", "F0E68C", "FFF8DC", "DAA520", "FFFAF0", "FDF5E6", "F5DEB3", "FFE4B5", "FFA500", "FFEFD5", "FFEBCD", "FFDEAD", "FAEBD7", "D2B48C", "DEB887", "FFE4C4", "FF8C00", "FAF0E6", "CD853F", "FFDAB9", "F4A460", "D2691E", "8B4513", "FFF5EE", "A0522D", "FFA07A", "FF7F50", "FF4500", "E9967A", "FF6347", "FFE4E1", "FA8072", "F08080", "BC8F8F", "CD5C5C", "FF0000", "A52A2A", "B22222", "8B0000", "800000", "F5F5F5", "DCDCDC", "D3D3D3", "C0C0C0", "A9A9A9", "808080", "696969"}

func getRandomColor() string {
	rand.Seed(time.Now().Unix())
	return colors[rand.Intn(len(colors))]
}
