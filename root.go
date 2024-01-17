package main

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var root = &cobra.Command{
	Use:   "netdog target",
	Short: "netdog is a reader/writer for TCP/unix socket",
	Long:  "read from stdin/files and write to <target> as is",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		network, err := getNetwork(cmd.Flags())
		if err != nil {
			return err
		}

		tlsConfig, err := getTlsConfig(cmd.Flags())
		if err != nil {
			return err
		}

		interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			return err
		}

		r := NewRunner()

		if len(args) > 1 {
			rds := make([]io.Reader, len(args)-1)
			for i := range args[1:] {
				rds[i], err = os.Open(args[i+1])
				if err != nil {
					log.Printf("invalid input: %q", err)
					break
				}
				if interval > 0 {
					rds[i] = newDelayedReader(rds[i], interval)
				}
			}
			r.Input = io.MultiReader(rds...)
		}

		return r.WriteToSocket(network, target, tlsConfig)
	},
}

func newDelayedReader(r io.Reader, delay time.Duration) io.Reader {
	return &delayedReader{r: r, delay: delay}
}

type delayedReader struct {
	r         io.Reader
	delay time.Duration
}

func (r *delayedReader) Read(p []byte) (int, error) {
	if r.delay > 0 {
		time.Sleep(r.delay)
		r.delay = 0 // only delay once
	}

	return r.r.Read(p)
}

func init() {
	root.PersistentFlags().Bool("unix-socket", false, "the target is unix socket path")

	root.Flags().Bool("tls", false, "dial using TLS")
	root.Flags().String("rootca", "", "root ca file")
	root.Flags().BoolP("insecure", "k", false, "skip TLS verification")
	root.Flags().DurationP("interval", "t", time.Duration(0), "interval between each message")

	root.AddCommand(dialCmd)
	root.AddCommand(lookupCmd)
}

func (r *Runner) WriteToSocket(network, target string, tlsConfig *tls.Config) error {
	var socket net.Conn
	var err error

	if tlsConfig != nil {
		socket, err = tls.Dial(network, target, tlsConfig)
	} else {
		socket, err = net.Dial(network, target)
	}

	if err != nil {
		return err
	}

	defer socket.Close()

	if _, err = io.Copy(socket, r.Input); err != nil {
		return err
	}

	if _, err = io.Copy(r.Output, socket); err != nil {
		return err
	}

	return nil
}

func getTlsConfig(flags *pflag.FlagSet) (*tls.Config, error) {
	if insecure, _ := flags.GetBool("insecure"); insecure {
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}

	if rootCa, _ := flags.GetString("rootca"); rootCa != "" {
		pem, err := os.ReadFile(rootCa)
		if err != nil {
			return nil, err
		}

		certs := x509.NewCertPool()
		certs.AppendCertsFromPEM(pem)

		return &tls.Config{
			RootCAs: certs,
		}, nil
	}

	if isTls, _ := flags.GetBool("tls"); isTls {
		return &tls.Config{}, nil
	} else {
		return nil, nil
	}
}

func getNetwork(flags *pflag.FlagSet) (string, error) {
	fUnixSocket, err := flags.GetBool("unix-socket")
	if err != nil {
		return "", err
	}

	if fUnixSocket {
		return "unix", nil
	} else {
		return "tcp", nil
	}
}
