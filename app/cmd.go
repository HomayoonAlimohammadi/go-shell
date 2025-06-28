package main

import "errors"

var (
	ErrNotFound = errors.New("command not found")
)

func handleCmd(cmd string) error {
	return ErrNotFound
}
