package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var dialCmd = &cobra.Command{
	Use:   "dial [target]",
	Short: "dial opens a tcp connection to the target and time the handshake",
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
		var tlsConfig *tls.Config

		isTls, err := cmd.Flags().GetBool("tls")
		if err != nil {
			return err
		}

		if isTls {
			rootCa, err := cmd.Flags().GetString("rootca")
			if err != nil {
				return err
			}

			var skipVerify bool
			var roots *x509.CertPool

			if rootCa == "skip" {
				skipVerify = true
			} else if rootCa != "" {
				pem, err := ioutil.ReadFile(rootCa)
				if err != nil {
					return err
				}

				roots = x509.NewCertPool()
				roots.AppendCertsFromPEM(pem)
			}

			tlsConfig = &tls.Config{
				InsecureSkipVerify: skipVerify,
				RootCAs:            roots,
			}
		}

		return r.Dial(network, target, tlsConfig)
	},
}

func init() {
	dialCmd.Flags().Bool("tls", false, "dial using TLS")
	dialCmd.Flags().String("rootca", "", "root ca file")
}

func (r Runner) Dial(network, target string, tlsConfig *tls.Config) error {
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
			var dialer net.Dialer = net.Dialer{
				Timeout: 5 * time.Second,
			}
			var err error
			var conn net.Conn

			startTm := time.Now()
			if tlsConfig != nil {
				conn, err = tls.DialWithDialer(&dialer, network, t, tlsConfig)
			} else {
				conn, err = dialer.Dial(network, t)
			}
			endTm := time.Now()
			if err != nil {
				r.ErrPrintfln("%s\terror connecting: %s", t, err)
				return
			}

			conn.Close()

			r.Printfln("%s\t%s", t, endTm.Sub(startTm))
		}()
	}

	wg.Wait()

	return nil
}
