package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

var (
	ErrNotFound = errors.New("command not found")

	builtins = map[string]struct{}{
		CmdExit: {},
		CmdEcho: {},
		CmdType: {},
	}
)

const (
	CmdExit = "exit"
	CmdEcho = "echo"
	CmdType = "type"
	CmdPwd  = "pwd"
)

func handleCmd(cmd string, args ...string) error {
	switch cmd {
	case CmdExit:
		return handleExit(args...)
	case CmdEcho:
		return handleEcho(args...)
	case CmdType:
		return handleType(args...)
	case CmdPwd:
		return handlePwd(args...)
	default:
		if path := searchPathFor(cmd); path != "" {
			return runExecutable(path, args...)
		}
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
	if len(args) == 0 {
		return errors.New("not enough arguments")
	}

	for _, cmd := range args {
		if _, ok := builtins[cmd]; ok {
			fmt.Printf("%s is a shell builtin\n", cmd)
			continue
		}

		if path := searchPathFor(cmd); path != "" {
			fmt.Printf("%s is %s\n", cmd, path)
			continue
		}

		fmt.Printf("%s: not found\n", cmd)
	}
	return nil
}

func handlePwd(args ...string) error {
	if len(args) > 0 {
		return errors.New("too many arguments")
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	fmt.Println(pwd)
	return nil
}

func searchPathFor(executable string) string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return ""
	}

	for dir := range strings.SplitSeq(pathEnv, ":") {
		relPath := path.Join(dir, executable)
		if info, err := os.Stat(relPath); err == nil {
			if info.Mode()&0111 != 0 {
				return relPath
			}
		}
	}
	return ""
}

func runExecutable(path string, args ...string) error {
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
