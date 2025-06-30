package main

import (
	"io"
	"sync"
)

type writer struct {
	sync.Mutex
	p []byte
}

var _ io.Writer = &writer{}

func (w *writer) Write(p []byte) (int, error) {
	w.Lock()
	w.p = append(w.p, p...)
	w.Unlock()
	return len(p), nil
}

func (w *writer) Content() []byte {
	out := make([]byte, len(w.p))
	copy(out, w.p)
	return out
}
