package main

import (
	"context"
	"io"
	"net"
	"regexp"
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

		withGrep, err := cmd.Flags().GetBool("grep")
		if err != nil {
			return err
		}

		var addr []string
		if len(args) > 0 {
			addr = args
		}

		return r.Lookup(addr, withName, withGrep)
	},
}

func init() {
	lookupCmd.Flags().BoolP("name", "n", false, "lookup names of the IP address(es)")
	lookupCmd.Flags().BoolP("grep", "g", false, "grep IP addresses from input using regex")
}

func (r *Runner) Lookup(args []string, withName bool, withGrep bool) error {
	var addresses []string
	if args != nil {
		addresses = args
	} else {
		addresses = r.ReadLines()
	}

	if withGrep {
	    // define a regex to match IPv4 addresses
		ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
		addresses = ipRegex.FindAllString(strings.Join(addresses, " "), -1)
	}

	if len(addresses) == 0 {
	    return nil
	}

	var wgAddresses sync.WaitGroup
	var resolver = &net.Resolver{
		PreferGo: true,
	}

	tw := tabwriter.NewWriter(r.Output, 0, 0, 4, ' ', 0)
	sw := newSyncWriter(tw)
	var lockAddr sync.Mutex

	for _, ad := range addresses {
		ad := ad
		wgAddresses.Add(1)
		go func() {
			defer wgAddresses.Done()
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

			lockAddr.Lock()
			defer lockAddr.Unlock()

			row{
				Host: ad,
				Time: endTm.Sub(startTm),
				IP:   result[0],
				Name: lookupName(result[0]),
			}.print(sw)

			var wgIP sync.WaitGroup
			for _, ip := range result[1:] {
				ip := ip
				wgIP.Add(1)
				go func() {
					defer wgIP.Done()
					row{
						Host: "",
						Time: 0,
						IP:   ip,
						Name: lookupName(ip),
					}.print(sw)
				}()
			}

			wgIP.Wait()
			Fprintfln(sw, "\t\t\t")
		}()
	}

	wgAddresses.Wait()
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
