package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

type Runner struct {
	Input     io.Reader
	Output    io.Writer
	ErrOutput io.Writer
	mux       *sync.Mutex
}

func NewRunner() *Runner {
	return &Runner{
		Input:     os.Stdin,
		Output:    os.Stdout,
		ErrOutput: os.Stderr,
		mux:       &sync.Mutex{},
	}
}

func (r *Runner) ReadLines() []string {
	lines := make([]string, 0)
	scanner := bufio.NewScanner(r.Input)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func (r *Runner) Printfln(format string, args ...interface{}) {
	r.mux.Lock()
	defer r.mux.Unlock()

	fmt.Fprintf(r.Output, format+"\n", args...)
}

func (r *Runner) ErrPrintfln(format string, args ...interface{}) {
	r.mux.Lock()
	defer r.mux.Unlock()

	fmt.Fprintf(r.ErrOutput, format+"\n", args...)
}
