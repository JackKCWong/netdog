package main

import (
	"context"
	"io"
	"net"
	"strings"
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
		withName, err := cmd.Flags().GetBool("name")
		if err != nil {
			return err
		}

		var addr string
		if len(args) > 0 {
			addr = args[0]
		}

		return r.Lookup(addr, withName)
	},
}

func init() {
	lookupCmd.Flags().Bool("name", false, "lookup names of the IP address(es)")
}

func (r *Runner) Lookup(addr string, withName bool) error {
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

	tw := tabwriter.NewWriter(r.Output, 0, 0, 4, ' ', 0)
	sw := newMutWriter(tw)
	var mu sync.Mutex

	for _, ad := range addresses {
		ad := ad
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			startTm := time.Now()
			result, err := resolver.LookupHost(ctx, ad)
			endTm := time.Now()

			if err != nil {
				r.ErrPrintfln("%s\t%s\terror resolving: %s", ad, endTm.Sub(startTm), err)
				return
			}

			lookupName := func(ip string) string {
				return ""
			}

			if withName {
				lookupName = func(ip string) string {
					names, err := resolver.LookupAddr(ctx, ip)
					if err != nil {
						return err.Error()
					}

					return strings.Join(names, " ")
				}
			}

			mu.Lock()
			defer mu.Unlock()

			row{
				Host: ad,
				Time: endTm.Sub(startTm),
				IP:   result[0],
				Name: lookupName(result[0]),
			}.print(sw)

			var ipWg sync.WaitGroup
			for _, ip := range result[1:] {
				ip := ip
				ipWg.Add(1)
				go func() {
					defer ipWg.Done()
					row{
						Host: "",
						Time: 0,
						IP:   ip,
						Name: lookupName(ip),
					}.print(sw)
				}()
			}

			ipWg.Wait()
			Fprintfln(sw, "\t\t\t")
		}()
	}

	wg.Wait()
	err := tw.Flush()
	if err != nil {
		return err
	}

	return nil
}

type row struct {
	Host string
	Time time.Duration
	IP   string
	Name string
}

func (r row) print(w io.Writer) {
	if r.Time == 0 {
		Fprintfln(w, "%s\t\t%s\t%s", r.Host, r.IP, r.Name)
	} else {
		Fprintfln(w, "%s\t%s\t%s\t%s", r.Host, r.Time, r.IP, r.Name)
	}
}
