package main

import "fmt"

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
