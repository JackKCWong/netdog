package main

import (
	"fmt"
	"io"
	"sync"
)

func Fprintfln(w io.Writer, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w, format+"\n", args...)
}

func newMutWriter(w io.Writer) *mutWriter {
	return &mutWriter{
		w: w,
	}
}

type mutWriter struct {
	m sync.Mutex
	w io.Writer
}

func (c *mutWriter) Write(p []byte) (n int, err error) {
	c.m.Lock()
	defer c.m.Unlock()

	return c.w.Write(p)
}
