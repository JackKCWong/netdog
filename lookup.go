package main

import (
	"context"
	"net"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:   "lookup [addr]",
	Short: "lookup address(es) to IP address(es) or names",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := NewRunner()

		if len(args) > 0 {
			return r.Lookup(args[0])
		} else {
			return r.Lookup("")
		}
	},
}

func (r Runner) Lookup(addr string) error {
	var addresses []string
	if addr != "" {
		addresses = []string{addr}
	} else {
		addresses = r.ReadLines()
	}

	var wg sync.WaitGroup
	var resolver = &net.Resolver{
		PreferGo: true,
	}

	tw := tabwriter.NewWriter(r.Output, 0, 0, 1, ' ', 0)
	var mu sync.Mutex

	for _, ad := range addresses {
		ad := ad
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			startTm := time.Now()
			result, err := resolver.LookupHost(ctx, ad)
			if err != nil {
				r.ErrPrintfln("%s\terror resolving: %s", ad, err)
				return
			}

			endTm := time.Now()

			mu.Lock()
			defer mu.Unlock()

			for i, ip := range result {
				if i == 0 {
					Fprintfln(tw, "%s\t%s\t%s\t", ad, endTm.Sub(startTm), ip)
				} else {
					Fprintfln(tw, "\t\t%s\t", ip)
				}
			}

			Fprintfln(tw, "\t\t\t")
		}()
	}

	wg.Wait()
	tw.Flush()
	return nil
}
