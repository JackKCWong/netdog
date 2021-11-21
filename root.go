package main

import (
	"io"
	"net"

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

		r := NewRunner()

		return r.WriteToSocket(network, target)
	},
}

func init() {
	root.PersistentFlags().Bool("unix-socket", false, "the target is unix socket path")

	root.AddCommand(fetchCmd)
	root.AddCommand(lookupCmd)
}

func (r Runner) WriteToSocket(network, target string) error {
	var socket net.Conn
	var err error
	socket, err = net.Dial(network, target)
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
