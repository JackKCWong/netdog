package main

import (
	"context"
	"net"
	"strings"
	"sync"
	"text/tabwriter"
	"text/template"
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

func (r Runner) Lookup(addr string, withName bool) error {
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
	var mu sync.Mutex
	firstT := template.Must(template.New("first").Parse(`{{.Host}}{{"\t"}}{{.Time}}{{"\t"}}{{.IP}}{{"\t"}}{{.Name}}{{"\n"}}`))
	restT := template.Must(template.New("rest").Parse(`{{"\t\t"}}{{.IP}}{{"\t"}}{{.Name}}{{"\n"}}`))

	for _, ad := range addresses {
		ad := ad
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			startTm := time.Now()
			result, err := resolver.LookupHost(ctx, ad)
			if err != nil {
				r.ErrPrintfln("%s\terror resolving: %s", ad, err)
				return
			}

			endTm := time.Now()

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

			firstT.Execute(tw, &row{
				Host: ad,
				Time: endTm.Sub(startTm),
				IP:   result[0],
				Name: lookupName(result[0]),
			})

			for _, ip := range result[1:] {
				restT.Execute(tw, &row{
					Host: "",
					Time: 0,
					IP:   ip,
					Name: lookupName(ip),
				})
			}

			Fprintfln(tw, "\t\t\t")
		}()
	}

	wg.Wait()
	tw.Flush()
	return nil
}

type row struct {
	Host string
	Time time.Duration
	IP   string
	Name string
}
