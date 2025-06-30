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
		CmdPwd:  {},
		CmdCd:   {},
	}
)

const (
	CmdExit = "exit"
	CmdEcho = "echo"
	CmdType = "type"
	CmdPwd  = "pwd"
	CmdCd   = "cd"
)

type redirOp int

const (
	RedirTypeUnknown redirOp = iota
	RedirOpRedir
	RedirOpAppend
)

type redirType int

const (
	RedirTypeNone redirType = iota
	RedirTypeBoth
	RedirTypeStdout
	RedirTypeStderr
)

type commandHandler struct {
	cmd       string
	args      []string
	redirType redirType
	redirOp   redirOp
	redirFile *os.File
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
}

func NewCommandHandler(parts []string, stdin io.Reader, stdout, stderr io.Writer) (*commandHandler, error) {
	var (
		typ  redirType
		op   redirOp
		dest string
	)

	if len(parts) >= 3 {
		p := parts[len(parts)-2]
		dest = parts[len(parts)-1]
		switch {
		case p == ">>":
			typ = RedirTypeStdout
			op = RedirOpAppend
			parts = parts[:len(parts)-2]
		case p == ">":
			typ = RedirTypeStdout
			op = RedirOpRedir
			parts = parts[:len(parts)-2]
		case len(p) == 3 && strings.HasSuffix(p, ">>"):
			switch string(p[0]) {
			case "&":
				typ = RedirTypeBoth
				op = RedirOpAppend
				parts = parts[:len(parts)-2]
			case "1":
				typ = RedirTypeStdout
				op = RedirOpAppend
				parts = parts[:len(parts)-2]
			case "2":
				typ = RedirTypeStderr
				op = RedirOpAppend
				parts = parts[:len(parts)-2]
			}
		case len(p) == 2 && strings.HasSuffix(p, ">"):
			switch string(p[0]) {
			case "&":
				typ = RedirTypeBoth
				op = RedirOpRedir
				parts = parts[:len(parts)-2]
			case "1":
				typ = RedirTypeStdout
				op = RedirOpRedir
				parts = parts[:len(parts)-2]
			case "2":
				typ = RedirTypeStderr
				op = RedirOpRedir
				parts = parts[:len(parts)-2]
			}
		}
	}

	var file *os.File
	switch op {
	case RedirOpRedir:
		var err error
		file, err = os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open %q for write: %w", dest, err)
		}
	case RedirOpAppend:
		var err error
		file, err = os.OpenFile(dest, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open %q for write: %w", dest, err)
		}
	}

	cmd, args := parts[0], parts[1:]

	return &commandHandler{
		cmd:       cmd,
		args:      args,
		redirType: typ,
		redirOp:   op,
		redirFile: file,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
	}, nil
}

func (r *commandHandler) StdoutWriter() io.Writer {
	if r.redirType == RedirTypeStdout || r.redirType == RedirTypeBoth {
		return r.redirFile
	}
	return r.stdout
}

func (r *commandHandler) StderrWriter() io.Writer {
	if r.redirType == RedirTypeStderr || r.redirType == RedirTypeBoth {
		return r.redirFile
	}
	return r.stderr
}

func (r *commandHandler) StdinReader() io.Reader {
	return r.stdin
}

func (r *commandHandler) Close() error {
	if r.redirFile != nil {
		return r.redirFile.Close()
	}
	return nil
}

func (h *commandHandler) Handle() error {
	if err := h.handleCmd(); err != nil {
		fmt.Fprintln(h.StderrWriter(), err)
	}

	return h.Close()
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
	cmd.Stdin = h.StdinReader()
	cmd.Run()
	return nil
}
