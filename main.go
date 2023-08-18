package main

import (
	"os"
	"runtime/debug"
)

func main() {
	debug.SetGCPercent(-1)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
