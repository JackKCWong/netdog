package main

import (
	"fmt"
	"io"
	"sync"
)

func Fprintfln(w io.Writer, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(w, format+"\n", args...)
}

func newSyncWriter(w io.Writer) *syncWriter {
	return &syncWriter{
		w: w,
	}
}

type syncWriter struct {
	m sync.Mutex
	w io.Writer
}

func (c *syncWriter) Write(p []byte) (n int, err error) {
	c.m.Lock()
	defer c.m.Unlock()

	return c.w.Write(p)
}
