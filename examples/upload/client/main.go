package main

import (
	"fmt"

	"github.com/fupengl/surf"
)

func main() {
	file := surf.NewMultipartFile(0)
	file.AddFields(map[string]string{
		"name":  "fupengl",
		"email": "fupenglxy@gmail.com",
	})
	file.AddFileFromPath("files", "README.md")
	resp, err := surf.Default.Upload("http://127.0.0.1:8888/upload", file)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Text())
}
