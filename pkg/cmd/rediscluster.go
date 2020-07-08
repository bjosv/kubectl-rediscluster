package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewRedisclusterCmd(streams genericclioptions.IOStreams) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "kubectl-rediscluster",
		Short: "Redis Cluster commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("a command is required")
		},
	}

	cmd.AddCommand(newVersionCmd(streams.Out))
	cmd.AddCommand(newSlotsCmd(streams))
	return cmd
}
