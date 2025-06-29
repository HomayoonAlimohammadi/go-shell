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

		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read command: %s\n", err)
		}

		if len(line) == 0 || line == "\n" {
			continue
		}

		line = line[:len(line)-1]

		parts := splitLine(line)
		// h, err := NewCommandHandler(parts)
		// if err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// }

		// if err := h.Handle(); err != nil {
		// 	fmt.Fprintln(os.Stderr, err)
		// }

		parts, redir := buildRedirector(parts)
		cmd, args := parts[0], parts[1:]

		if err := handleCmd(redir.StdoutWriter(), redir.StderrWriter(), cmd, args...); err != nil {
			fmt.Fprintln(redir.StderrWriter(), err)
		}

		redir.Close()
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
		} else if b == '"' && !inSingleQuote {
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

	if len(part) > 0 {
		parts = append(parts, part)
	}

	return parts
}
