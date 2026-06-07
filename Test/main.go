package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var url string
	fmt.Println("testing github clone")
	fmt.Println("Enter Url")
	fmt.Scanln(&url)

	cmd := exec.Command("git", "clone", url)
	output , err := cmd.CombinedOutput()
	if err !=  nil {
		fmt.Println(err)
		fmt.Println(string(output))
	}
	fmt.Println("Done")
}