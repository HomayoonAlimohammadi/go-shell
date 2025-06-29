package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
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
	CmdCd   = "cd"
)

func handleCmd(stdout io.Writer, stderr io.Writer, cmd string, args ...string) error {
	switch cmd {
	case CmdExit:
		return handleExit(stdout, args...)
	case CmdEcho:
		return handleEcho(stdout, args...)
	case CmdType:
		return handleType(stdout, args...)
	case CmdPwd:
		return handlePwd(stdout, args...)
	case CmdCd:
		return handleCd(stdout, args...)
	default:
		if path := searchPathFor(cmd); path != "" {
			return runExecutable(stdout, stderr, path, args...)
		}
		return ErrNotFound
	}
}

func handleExit(out io.Writer, args ...string) error {
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

func handleEcho(out io.Writer, args ...string) error {
	fmt.Fprintln(out, strings.Join(args, " "))
	return nil
}

func handleType(out io.Writer, args ...string) error {
	if len(args) == 0 {
		return errors.New("not enough arguments")
	}

	for _, cmd := range args {
		if _, ok := builtins[cmd]; ok {
			fmt.Fprintf(out, "%s is a shell builtin\n", cmd)
			continue
		}

		if path := searchPathFor(cmd); path != "" {
			fmt.Fprintf(out, "%s is %s\n", cmd, path)
			continue
		}

		fmt.Fprintf(out, "%s: not found\n", cmd)
	}
	return nil
}

func handlePwd(out io.Writer, args ...string) error {
	if len(args) > 0 {
		return errors.New("too many arguments")
	}

	wd, _ := os.Getwd()
	fmt.Fprintln(out, wd)
	return nil
}

func handleCd(out io.Writer, args ...string) error {
	if len(args) > 1 {
		return errors.New("too many arguments")
	}

	if len(args) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		return syscall.Chdir(homeDir)
	}

	dir := args[0]
	if dir == "~" {
		dir = os.Getenv("HOME")
	}

	return syscall.Chdir(dir)
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

func runExecutable(stdout io.Writer, stderr io.Writer, path string, args ...string) error {
	cmd := exec.Command(path, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Run()
	return nil
}
