package main

import (
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping [target]",
	Short: "ping the targets and time the tcp & tls handshake",
	Long:  "target can be specified as arg (single) or from stdin (one line per target), in the format of host:port or /path/to/unixsocket.",
	RunE: func(cmd *cobra.Command, args []string) error {
		network, err := getNetwork(cmd.Flags())
		if err != nil {
			return err
		}

		r := NewRunner()
		tw := tabwriter.NewWriter(r.Output, 0, 0, 4, ' ', 0)
		r.Output = tw
		defer tw.Flush()

		var targets []string
		if len(args) > 0 {
			targets = args
		} else {
			targets = r.ReadLines()
		}

		tlsConfig, err := getTlsConfig(cmd.Flags())
		if err != nil {
			return err
		}

		sniff, _ := cmd.Flags().GetBool("sniff")

		var wg sync.WaitGroup
		for i := range targets {
			wg.Add(1)
			go func(target string) {
				defer wg.Done()
				for range time.Tick(1 * time.Second) {
					if err := r.Dial(network, target, tlsConfig, sniff); err != nil {
						r.ErrPrintfln("%s", err)
					}
					tw.Flush()
				}
			}(targets[i])
		}

		wg.Wait()
		return nil
	},
}

func init() {
	pingCmd.Flags().Bool("tls", false, "dial using TLS")
	pingCmd.Flags().Bool("sniff", false, "dial to all the IPs behind a DNS")
	pingCmd.Flags().String("rootca", "", "root ca file")
	pingCmd.Flags().BoolP("insecure", "k", false, "skip TLS verification")
}
