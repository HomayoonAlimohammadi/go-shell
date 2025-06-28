package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Print("$ ")

	cmd, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Printf("failed to read command: %s", err)
	}

	cmd = cmd[:len(cmd)-1]
	if err := handleCmd(cmd); err != nil {
		fmt.Printf("%s: %s", cmd, err)
	}
}
