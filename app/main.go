package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	for {
		printContext()

		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read command: %s\n", err)
		}

		if len(line) == 0 || line == "\n" {
			continue
		}

		line = line[:len(line)-1]

		pipedCmds := splitCmds(line)

		var stdin io.Reader = os.Stdin
		for i, cmd := range pipedCmds {
			stdout := &writer{}
			stderr := &writer{}

			h, err := NewCommandHandler(cmd, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(stderr, err)
			}

			if err := h.Handle(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			stdin = bytes.NewReader(stdout.Content())
			if len(pipedCmds) == 1 || i == len(pipedCmds)-1 {
				os.Stdout.Write(stdout.Content())
			}
			os.Stderr.Write(stderr.Content())
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

func splitCmds(l string) [][]string {
	var (
		pipedCmds     [][]string
		cmdParts      []string
		cmdPart       string
		inSingleQuote bool
		inDoubleQuote bool
	)
	for _, b := range l {
		if string(b) == "'" && !inDoubleQuote {
			if inSingleQuote {
				inSingleQuote = false
				cmdParts = append(cmdParts, cmdPart)
				cmdPart = ""
			} else {
				inSingleQuote = true
			}
		} else if b == '"' && !inSingleQuote {
			if inDoubleQuote {
				inDoubleQuote = false
				cmdParts = append(cmdParts, cmdPart)
				cmdPart = ""
			} else {
				inDoubleQuote = true
			}
		} else if b == ' ' && !inSingleQuote && !inDoubleQuote {
			if len(cmdPart) > 0 {
				cmdParts = append(cmdParts, cmdPart)
				cmdPart = ""
			}
		} else if b == '|' && !inSingleQuote && !inDoubleQuote {
			if len(cmdPart) > 0 {
				cmdParts = append(cmdParts, cmdPart)
				cmdPart = ""
			}
			pipedCmds = append(pipedCmds, cmdParts)
			cmdParts = []string{}
		} else {
			cmdPart += string(b)
		}
	}

	if len(cmdPart) > 0 {
		cmdParts = append(cmdParts, cmdPart)
	}

	if len(cmdParts) > 0 {
		pipedCmds = append(pipedCmds, cmdParts)
	}

	return pipedCmds
}
