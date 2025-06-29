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
		parts := splitLine(cmdLine)
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

func splitLine(l string) []string {
	var (
		parts         []string
		part          string
		inSingleQuote bool
		inDoubleQuote bool
	)
	for _, b := range l {
		if string(b) == "'" && !inDoubleQuote {
			if inSingleQuote {
				inSingleQuote = false
				parts = append(parts, part)
				part = ""
			} else {
				inSingleQuote = true
			}
		} else if string(b) == "\"" && !inSingleQuote {
			if inDoubleQuote {
				inDoubleQuote = false
				parts = append(parts, part)
				part = ""
			} else {
				inDoubleQuote = true
			}
		} else if b == ' ' && !inSingleQuote && !inDoubleQuote {
			if len(part) > 0 {
				parts = append(parts, part)
				part = ""
			}
		} else {
			part += string(b)
		}
	}
	return parts
}
