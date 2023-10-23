package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"sync"
	"text/tabwriter"
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
		tw := tabwriter.NewWriter(r.Output, 0, 0, 4, ' ', 0)
		r.Output = tw

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

		err = r.Dial(network, target, tlsConfig)
		if err != nil {
			return err
		}
		return tw.Flush()
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
			var connErr error
			var conn net.Conn
			var tlsConn *tls.Conn
			var tlsVersions = map[uint16]string{
				tls.VersionTLS10: "TLS1.0",
				tls.VersionTLS11: "TLS1.1",
				tls.VersionTLS12: "TLS1.2",
				tls.VersionTLS13: "TLS1.3",
			}

			startTm := time.Now()
			if tlsConfig != nil {
				tlsConn, connErr = tls.DialWithDialer(&dialer, network, t, tlsConfig)
				conn = tlsConn
			} else {
				conn, connErr = dialer.Dial(network, t)
			}
			if connErr != nil {
				r.ErrPrintfln("%s\terror connecting: %s", t, connErr)
				return
			}

			endTm := time.Now()
			defer conn.Close()
			ip := conn.RemoteAddr().String()
			if tlsConn == nil {
				r.Printfln("%s\t%s\t%s", t, endTm.Sub(startTm), ip)
			} else {
				state := tlsConn.ConnectionState()
				cipher := tls.CipherSuiteName(state.CipherSuite)
				tlsVer := tlsVersions[state.Version]
				sni := state.ServerName
				alpn := state.NegotiatedProtocol

				r.Printfln("%s\t%s\t%s\t%s\t%s\t%s\t%s", t, endTm.Sub(startTm),
					ip, sni, tlsVer, alpn, cipher)
			}
		}()
	}

	wg.Wait()

	return nil
}
