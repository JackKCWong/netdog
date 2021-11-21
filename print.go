package main

import (
	"fmt"
	"io"
)

func Fprintfln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}
