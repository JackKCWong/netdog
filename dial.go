package main

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
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

		for i := range targets {
			if err := r.Dial(network, targets[i], tlsConfig, sniff); err != nil {
				r.ErrPrintfln("%s", err)
			}
		}

		return nil
	},
}

func init() {
	dialCmd.Flags().Bool("tls", false, "dial using TLS")
	dialCmd.Flags().Bool("sniff", false, "dial to all the IPs behind a DNS")
	dialCmd.Flags().String("rootca", "", "root ca file")
	dialCmd.Flags().BoolP("insecure", "k", false, "skip TLS verification")
}

func (r *Runner) Dial(network string, target string, tlsConfig *tls.Config, sniff bool) error {
	resolver := &net.Resolver{
		PreferGo: true,
	}

	var addr []string
	parts := strings.Split(target, ":")
	serverName := parts[0]
	port := parts[1]
	if sniff {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		hosts, err := resolver.LookupHost(ctx, serverName)
		if err != nil {
			return err
		}

		for j := range hosts {
			addr = append(addr, hosts[j]+":"+port)
		}
	} else {
		addr = append(addr, target)
	}

	var wg sync.WaitGroup
	for _, adr := range addr {
		wg.Add(1)
		adr := adr
		go func() {
			defer wg.Done()
			if tlsConfig == nil {
				r.dialTcp(network, target, adr)
			} else {
				r.dialTls(network, target, adr, tlsConfig)
			}
		}()
	}

	wg.Wait()

	return nil
}

func (r *Runner) dialTcp(network, target, addr string) error {
	resolver := &net.Resolver{
		PreferGo: true,
	}

	var dialer net.Dialer = net.Dialer{
		Timeout: 5 * time.Second,
	}
	var conn net.Conn

	var tcpStartTm = time.Now()
	var tcpEndTm time.Time

	conn, connErr := dialer.Dial(network, addr)
	tcpEndTm = time.Now()

	if connErr != nil {
		return connErr
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
	names, _ := resolver.LookupAddr(ctx, ip) // just let it print empty string if error

	r.Printfln("%s\ttcp:%s\t%s\t%s", target, tcpEndTm.Sub(tcpStartTm), ip, strings.Join(names, ", "))

	return nil
}

func (r *Runner) dialTls(network, target, addr string, tlsConfig *tls.Config) error {
	resolver := &net.Resolver{
		PreferGo: true,
	}

	var dialer net.Dialer = net.Dialer{
		Timeout: 5 * time.Second,
	}

	var connErr error
	var tcpConn net.Conn
	var tlsConn *tls.Conn
	var tlsVersions = map[uint16]string{
		tls.VersionTLS10: "TLS1.0",
		tls.VersionTLS11: "TLS1.1",
		tls.VersionTLS12: "TLS1.2",
		tls.VersionTLS13: "TLS1.3",
	}

	var tcpStartTm = time.Now()
	var tcpEndTm time.Time
	var tlsEndTm time.Time

	tcpConn, connErr = dialer.Dial(network, addr)
	tcpEndTm = time.Now()

	if connErr != nil {
		return connErr
	}

	tlsConfig.ServerName = strings.Split(target, ":")[0]
	tlsConn = tls.Client(tcpConn, tlsConfig)
	defer tlsConn.Close()

	tlsConn.Handshake()
	tlsEndTm = time.Now()

	ip := strings.Split(tcpConn.RemoteAddr().String(), ":")[0]
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	names, _ := resolver.LookupAddr(ctx, ip)
	state := tlsConn.ConnectionState()
	cipher := tls.CipherSuiteName(state.CipherSuite)
	tlsVer := tlsVersions[state.Version]
	alpn := state.NegotiatedProtocol

	r.Printfln("%s\ttcp:%s\ttls:%s\ttotal:%s\t%s\t%s\t%s\t%s\t%s", target,
		tcpEndTm.Sub(tcpStartTm),
		tlsEndTm.Sub(tcpEndTm),
		tlsEndTm.Sub(tcpStartTm),
		ip, tlsVer, alpn, cipher, strings.Join(names, ", "))

	return nil
}
