package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		printContext()

		cmdLine, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Printf("failed to read command: %s\n", err)
		}

		if len(cmdLine) == 0 || cmdLine == "\n" {
			continue
		}

		cmdLine = cmdLine[:len(cmdLine)-1]
		parts := strings.Split(cmdLine, " ")
		cmd, args := parts[0], parts[1:]
		if err := handleCmd(cmd, args...); err != nil {
			fmt.Printf("%s: %s\n", cmdLine, err)
		}
	}
}

func printContext() {
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}

	wd, _ := os.Getwd()
	hd := os.Getenv("HOME")
	if strings.HasPrefix(wd, hd) {
		wd = "~" + wd[len(hd):]
	}

	fmt.Printf("%s@go-shell:%s$ ", user, wd)
}
