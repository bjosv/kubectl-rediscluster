package cmd

import (
	"errors"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type redisclusterCmd struct {
	out io.Writer
}

func NewRedisclusterCmd(streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "rediscluster",
		Short: "Redis Cluster commands",
		// SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("no arguments accepted for this command")
			}
			return nil
		},
	}

	cmd.AddCommand(newVersionCmd(streams.Out))
	return cmd
}
