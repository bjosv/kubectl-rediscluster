package main

import (
	"os"

	"github.com/bjosv/kubectl-rediscluster/pkg/cmd"
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

	root := cmd.NewRedisclusterCmd(
		genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
