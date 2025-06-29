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

type commandHandler struct {
	cmd   string
	args  []string
	redir *redirector
}

func NewCommandHandler(parts []string, redir *redirector) *commandHandler {
	cmd, args := parts[0], parts[1:]

	return &commandHandler{
		cmd:   cmd,
		args:  args,
		redir: redir,
	}
}

func (h *commandHandler) Handle() error {
	if err := h.handleCmd(); err != nil {
		fmt.Fprintln(h.StderrWriter(), err)
	}

	return h.redir.Close()
}

func (h *commandHandler) StdoutWriter() io.Writer {
	return h.redir.StdoutWriter()
}

func (h *commandHandler) StderrWriter() io.Writer {
	return h.redir.StderrWriter()
}

func (h *commandHandler) handleCmd() error {
	switch h.cmd {
	case CmdExit:
		return h.handleExit()
	case CmdEcho:
		return h.handleEcho()
	case CmdType:
		return h.handleType()
	case CmdPwd:
		return h.handlePwd()
	case CmdCd:
		return h.handleCd()
	default:
		if path := searchPathFor(h.cmd); path != "" {
			return h.runExecutable(path)
		}
		return ErrNotFound
	}
}

func (h *commandHandler) handleExit() error {
	code := 0
	if len(h.args) > 0 {
		var err error
		if code, err = strconv.Atoi(h.args[0]); err != nil {
			return fmt.Errorf("invalid code %s: %e", h.args[0], err)
		}
	}
	os.Exit(code)
	return nil
}

func (h *commandHandler) handleEcho() error {
	fmt.Fprintln(h.StdoutWriter(), strings.Join(h.args, " "))
	return nil
}

func (h *commandHandler) handleType() error {
	if len(h.args) == 0 {
		return errors.New("not enough arguments")
	}

	for _, cmd := range h.args {
		if _, ok := builtins[cmd]; ok {
			fmt.Fprintf(h.StdoutWriter(), "%s is a shell builtin\n", cmd)
			continue
		}

		if path := searchPathFor(cmd); path != "" {
			fmt.Fprintf(h.StdoutWriter(), "%s is %s\n", cmd, path)
			continue
		}

		fmt.Fprintf(h.StderrWriter(), "%s: not found\n", cmd)
	}
	return nil
}

func (h *commandHandler) handlePwd() error {
	if len(h.args) > 0 {
		return errors.New("too many arguments")
	}

	wd, _ := os.Getwd()
	fmt.Fprintln(h.StdoutWriter(), wd)
	return nil
}

func (h *commandHandler) handleCd() error {
	if len(h.args) > 1 {
		return errors.New("too many arguments")
	}

	if len(h.args) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		return syscall.Chdir(homeDir)
	}

	dir := h.args[0]
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

func (h *commandHandler) runExecutable(path string) error {
	cmd := exec.Command(path, h.args...)
	cmd.Stdout = h.StdoutWriter()
	cmd.Stderr = h.StderrWriter()
	cmd.Run()
	return nil
}
