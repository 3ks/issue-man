package client

import (
	"context"
	c "github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

var client *c.Client

// 单例模式
func Get() *c.Client {
	return client
}

// 初始化 GitHub Client
func Init(token string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "token"},
	)
	tc := oauth2.NewClient(ctx, ts)
	client = c.NewClient(tc)
}
