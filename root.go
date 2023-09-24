package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var root = &cobra.Command{
	Use:   "netdog target",
	Short: "netdog is a reader/writer for TCP/unix socket",
	Long:  "the default behavior is to read from stdin and write to <target>",
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

		delim, err := cmd.Flags().GetString("sep")
		if err != nil {
			return err
		}

		r := NewRunner()

		return r.WriteToSocket(network, target, delim, tlsConfig)
	},
}

func init() {
	root.PersistentFlags().Bool("unix-socket", false, "the target is unix socket path")

	root.Flags().Bool("tls", false, "dial using TLS")
	root.Flags().String("rootca", "", "root ca file")
	root.Flags().String("sep", "", "a message separator, useful when you want to send multiple 'message' to a message oriented protocol, like websocket")

	root.AddCommand(dialCmd)
	root.AddCommand(lookupCmd)
}

func (r Runner) WriteToSocket(network, target, delim string, tlsConfig *tls.Config) error {
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

	if len(delim) == 0 {
		if _, err = io.Copy(socket, r.Input); err != nil {
			return err
		}

		if _, err = io.Copy(r.Output, socket); err != nil {
			return err
		}
	} else {
		scanner := bufio.NewScanner(r.Input)
		scanner.Split(makeSplitFunc(delim))

		for scanner.Scan() {
			msg := scanner.Bytes()
			if _, err = socket.Write(msg); err != nil {
				return err
			}

		}

		if _, err = io.Copy(r.Output, socket); err != nil {
			return err
		}
	}

	return nil
}

func getTlsConfig(flags *pflag.FlagSet) (*tls.Config, error) {
	if isTls, err := flags.GetBool("tls"); err != nil {
		return nil, err
	} else {
		if !isTls {
			return nil, nil
		}
	}

	if rootCa, err := flags.GetString("rootca"); err != nil {
		return nil, err
	} else {
		if rootCa == "skip" {
			return &tls.Config{
				InsecureSkipVerify: true,
			}, nil
		} else if rootCa != "" {
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
	}

	return &tls.Config{}, nil
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

func makeSplitFunc(delim string) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := strings.Index(string(data), delim); i >= 0 {
			return i + len(delim), data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return
	}
}
