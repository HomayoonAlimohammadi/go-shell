package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Print("$ ")

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
