package main

import (
	"net"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch [target]",
	Short: "fetch opens a tcp connection to the target and time the handshake",
	Long:  "target can be specified as arg (single) or from stdin (one line per target), in the format of host:port or /path/to/unixsocket.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var target string
		if len(args) > 0 {
			target = args[0]
		}

		network, err := getNetwork(cmd.Flags())
		if err != nil {
			return err
		}

		r := NewRunner()

		return r.Fetch(network, target)
	},
}

func (r Runner) Fetch(network, target string) error {
	var targets []string
	if target != "" {
		targets = []string{target}
	} else {
		// read targets from stdin
		targets = r.ReadLines()
	}

	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		t := t
		go func() {
			defer wg.Done()
			startTm := time.Now()
			conn, err := net.DialTimeout(network, t, 10*time.Second)
			if err != nil {
				r.ErrPrintlnf("%s\terror connecting: %s", t, err)
				return
			}

			endTm := time.Now()
			conn.Close()

			r.Printlnf("%s\t%s", t, endTm.Sub(startTm))
		}()
	}

	wg.Wait()

	return nil
}
