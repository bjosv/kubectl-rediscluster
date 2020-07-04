package main

import (
	"os"

	"github.com/bjosv/kubectl-rediscluster/cmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Version of this plugin
var version = "undefined"

func main() {
	cmd.SetVersion(version)

	redisclusterCmd := cmd.NewRedisclusterCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := redisclusterCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
