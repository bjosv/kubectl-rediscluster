package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var version string
var commit string

func SetVersion(v string, c string) {
	version = v
	commit = c
}

type versionCmd struct {
	out io.Writer
}

func NewVersionCmd(out io.Writer) *cobra.Command {
	version := &versionCmd{out}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "plugin version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.run()
		},
	}
	return cmd
}

func (v *versionCmd) run() error {
	_, err := fmt.Fprintf(v.out, "Version:\t%s\nCommit:\t%s\n",
		version, commit)
	if err != nil {
		return err
	}
	return nil
}
