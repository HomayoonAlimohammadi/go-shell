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
	typ redirType
	op  redirOp
	f   *os.File
}

func NewRedirector(typ redirType, op redirOp, dest string) (*redirector, error) {
	var file *os.File
	switch op {
	case RedirOpRedir:
		var err error
		file, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
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
		typ: typ,
		op:  op,
		f:   file,
	}, nil
}

func NewStdRedirector() *redirector {
	return &redirector{}
}

func (r *redirector) StdoutWriter() io.Writer {
	if r.typ == RedirTypeStdout || r.typ == RedirTypeBoth {
		return r.f
	}
	return os.Stdout
}

func (r *redirector) StderrWriter() io.Writer {
	if r.typ == RedirTypeStderr || r.typ == RedirTypeBoth {
		return r.f
	}
	return os.Stderr
}

func (r *redirector) Close() {
	r.f.Close()
}

func buildRedirector(parts []string) ([]string, *redirector) {
	if len(parts) < 3 {
		return parts, NewStdRedirector()
	}

	p := parts[len(parts)-2]
	dest := parts[len(parts)-1]
	switch {
	case p == ">>":
		r, err := NewRedirector(RedirTypeStdout, RedirOpAppend, dest)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return parts[:len(parts)-2], r
	case p == ">":
		r, err := NewRedirector(RedirTypeStdout, RedirOpRedir, dest)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return parts[:len(parts)-2], r
	case len(p) == 3 && strings.HasSuffix(p, ">>"):
		switch string(p[0]) {
		case "&":
			r, err := NewRedirector(RedirTypeBoth, RedirOpAppend, dest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "1":
			r, err := NewRedirector(RedirTypeStdout, RedirOpAppend, dest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "2":
			r, err := NewRedirector(RedirTypeStderr, RedirOpAppend, dest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		}
	case len(p) == 2 && strings.HasSuffix(p, ">"):
		switch string(p[0]) {
		case "&":
			r, err := NewRedirector(RedirTypeBoth, RedirOpRedir, dest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "1":
			r, err := NewRedirector(RedirTypeStdout, RedirOpRedir, dest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		case "2":
			r, err := NewRedirector(RedirTypeStderr, RedirOpRedir, dest)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return parts[:len(parts)-2], r
		}
	}

	return parts, NewStdRedirector()
}
