package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrNotFound = errors.New("command not found")

	builtins = map[string]struct{}{
		CmdExit: struct{}{},
		CmdEcho: struct{}{},
		CmdType: struct{}{},
	}
)

const (
	CmdExit = "exit"
	CmdEcho = "echo"
	CmdType = "type"
)

func handleCmd(cmd string, args ...string) error {
	switch cmd {
	case CmdExit:
		return handleExit(args...)
	case CmdEcho:
		return handleEcho(args...)
	case CmdType:
		return handleType(args...)
	default:
		return ErrNotFound
	}
}

func handleExit(args ...string) error {
	code := 0
	if len(args) > 0 {
		var err error
		if code, err = strconv.Atoi(args[0]); err != nil {
			return fmt.Errorf("invalid code %s: %e", args[0], err)
		}
	}
	os.Exit(code)
	return nil
}

func handleEcho(args ...string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

func handleType(args ...string) error {
	for _, cmd := range args {
		if _, ok := builtins[cmd]; ok {
			fmt.Printf("%s is a shell builtin\n", cmd)
		} else {
			fmt.Printf("%s: not found\n", cmd)
		}
	}
	return nil
}
