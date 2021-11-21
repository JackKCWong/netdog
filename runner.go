package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Runner struct {
	Input     io.Reader
	Output    io.Writer
	ErrOutput io.Writer
}

func NewRunner() Runner {
	return Runner{
		Input:     os.Stdin,
		Output:    os.Stdout,
		ErrOutput: os.Stderr,
	}
}

func (r Runner) ReadLines() []string {
	lines := make([]string, 0)
	scanner := bufio.NewScanner(r.Input)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func (r Runner) Printlnf(format string, args ...interface{}) {
	fmt.Fprintf(r.Output, format+"\n", args...)
}

func (r Runner) ErrPrintlnf(format string, args ...interface{}) {
	fmt.Fprintf(r.ErrOutput, format+"\n", args...)
}
