// client 包，指的是 GitHub 客户端库的初始化和使用。
package client

import (
	"context"
	c "github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

var client *c.Client

// 需要先初始化，才能正常使用
func Get() *c.Client {
	return client
}

// 初始化 GitHub Client
func Init(token string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = c.NewClient(tc)
}
