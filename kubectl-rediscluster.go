package main

import (
	"os"

	"github.com/bjosv/kubectl-rediscluster/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Version of this plugin
var version = "undefined"

func main() {
	cmd.SetVersion(version)

	flags := pflag.NewFlagSet("kubectl-rediscluster", pflag.ExitOnError)
	pflag.CommandLine = flags

	redisclusterCmd := cmd.NewRedisclusterCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	flags.AddFlagSet(redisclusterCmd.PersistentFlags())

	// Add flags for kubectl configs:
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(flags)

	if err := redisclusterCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
