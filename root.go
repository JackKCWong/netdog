package main

import (
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "netdog target [port]",
	Short: "netdog is a reader/writer for TCP/unix socket",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var fUnixSocket bool
		var err error
		var target string

		if len(args) == 1 {
			fUnixSocket, err = cmd.Flags().GetBool("unix-socket")
			if err != nil {
				return err
			}

			target = args[0]
		}

		if len(args) == 2 {
			target = args[0] + ":" + args[1]
		}

		r := Runner{
			Input:  os.Stdin,
			Output: os.Stdout,
		}

		return r.Run(RunnerArgs{
			Target:       target,
			IsUnixSocket: fUnixSocket,
		})
	},
}

func init() {
	root.Flags().Bool("unix-socket", false, "the target is unix socket path")
}

type Runner struct {
	Input  io.Reader
	Output io.Writer
}

type RunnerArgs struct {
	Target       string
	IsUnixSocket bool
}

func (r Runner) Run(args RunnerArgs) error {
	var socket net.Conn
	var err error
	if args.IsUnixSocket {
		socket, err = net.Dial("unix", args.Target)
		if err != nil {
			return err
		}
	} else {
		socket, err = net.Dial("tcp", args.Target)
		if err != nil {
			return err
		}
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
