package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var (
	ErrNotFound = errors.New("command not found")
)

const (
	CmdExit = "exit"
)

func handleCmd(cmd string, args ...string) error {
	switch cmd {
	case CmdExit:
		return handleExit(args...)
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
