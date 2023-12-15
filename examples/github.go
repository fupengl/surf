package main

import (
	"fmt"

	"github.com/fupengl/surf"
)

type GithubApi struct {
	client *surf.Surf
}

func NewGithubApi() *GithubApi {
	config := surf.Config{
		BaseURL: "https://api.github.com",
		Headers: map[string][]string{
			"Accept":               {"application/vnd.github+json"},
			"X-GitHub-Api-Version": {"2022-11-28"},
		},
	}
	return &GithubApi{
		client: surf.New(&config),
	}
}

type GithubUserInfo struct {
	Login     string `json:"login"`
	Id        int    `json:"id"`
	AvatarUrl string `json:"avatar_url"`
}

func (api *GithubApi) GetUser(usrName string) (info *GithubUserInfo, err error) {
	var resp *surf.Response
	resp, err = api.client.Get("users/:username",
		surf.WithSetParam("username", usrName),
	)
	if err != nil {
		return
	}
	return info, resp.Json(&info)
}

type GithubRepoInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (api *GithubApi) GetRepo(OWNER, REPO string) (info *GithubRepoInfo, err error) {
	var resp *surf.Response
	resp, err = api.client.Get("repos/:OWNER/:REPO",
		surf.WithSetParam("OWNER", OWNER),
		surf.WithSetParam("REPO", REPO),
	)
	if err != nil {
		return
	}
	fmt.Printf("GetRepo perf: %+v\n", resp.Performance)
	return info, resp.Json(&info)
}

func main() {
	api := NewGithubApi()

	userInfo, err := api.GetUser("fupengl")
	if err != nil {
		panic(fmt.Errorf("get user info error %w", err))
	}
	fmt.Printf("github user fupengl info: %+v\n", userInfo)

	repoInfo, err := api.GetRepo("fupengl", "surf")
	if err != nil {
		panic(fmt.Errorf("get repo info error %w", err))
	}
	fmt.Printf("github repo info: %+v\n", repoInfo)
}
