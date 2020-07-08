package main

import (
	"os"

	"github.com/bjosv/kubectl-rediscluster/pkg/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Versions set at link time
var version = "undefined"
var git = "undefined"
var goversion = "undefined"

func main() {
	cmd.SetVersion(version, git, goversion)

	// Set new flagset name
	flags := pflag.NewFlagSet("kubectl-rediscluster", pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	root := &cobra.Command{
		Use:   "kubectl-rediscluster",
		Short: "A kubectl plugin for inspecting your Redis Cluster",
	}

	root.AddCommand(cmd.NewVersionCmd(streams.Out))
	root.AddCommand(cmd.NewSlotsCmd(streams))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
