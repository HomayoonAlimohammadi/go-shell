package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type redirOp string
type redirType string

const (
	RedirOpRedir  redirOp = "redir"
	RedirOpAppend redirOp = "append"

	RedirTypeBoth   redirType = "both"
	RedirTypeStdout redirType = "stdout"
	RedirTypeStderr redirType = "stderr"
)

type redirector struct {
	typ    redirType
	op     redirOp
	f      *os.File
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewRedirector(typ redirType, op redirOp, dest string, stdin io.Reader, stdout, stderr io.Writer) (*redirector, error) {
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
	default:
		return nil, fmt.Errorf("invalid operation %q", op)
	}

	return &redirector{
		typ:    typ,
		op:     op,
		f:      file,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func NewStdRedirector(stdin io.Reader, stdout, stderr io.Writer) *redirector {
	return &redirector{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

func (r *redirector) StdoutWriter() io.Writer {
	if r.typ == RedirTypeStdout || r.typ == RedirTypeBoth {
		return r.f
	}
	return r.stdout
}

func (r *redirector) StderrWriter() io.Writer {
	if r.typ == RedirTypeStderr || r.typ == RedirTypeBoth {
		return r.f
	}
	return r.stderr
}

func (r *redirector) StdinReader() io.Reader {
	return r.stdin
}

func (r *redirector) Close() error {
	if r.f != nil {
		return r.f.Close()
	}
	return nil
}

func splitPartsAndRedir(parts []string, stdin io.Reader, stdout, stderr io.Writer) ([]string, *redirector) {
	if len(parts) < 3 {
		return parts, NewStdRedirector(stdin, stdout, stderr)
	}

	p := parts[len(parts)-2]
	dest := parts[len(parts)-1]
	switch {
	case p == ">>":
		r, err := NewRedirector(RedirTypeStdout, RedirOpAppend, dest, stdin, stdout, stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return parts[:len(parts)-2], r
	case p == ">":
		r, err := NewRedirector(RedirTypeStdout, RedirOpRedir, dest, stdin, stdout, stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return parts[:len(parts)-2], r
	case len(p) == 3 && strings.HasSuffix(p, ">>"):
		switch string(p[0]) {
		case "&":
			r, err := NewRedirector(RedirTypeBoth, RedirOpAppend, dest, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "1":
			r, err := NewRedirector(RedirTypeStdout, RedirOpAppend, dest, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "2":
			r, err := NewRedirector(RedirTypeStderr, RedirOpAppend, dest, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		}
	case len(p) == 2 && strings.HasSuffix(p, ">"):
		switch string(p[0]) {
		case "&":
			r, err := NewRedirector(RedirTypeBoth, RedirOpRedir, dest, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "1":
			r, err := NewRedirector(RedirTypeStdout, RedirOpRedir, dest, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "2":
			r, err := NewRedirector(RedirTypeStderr, RedirOpRedir, dest, stdin, stdout, stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		}
	}

	return parts, NewStdRedirector(stdin, stdout, stderr)
}
