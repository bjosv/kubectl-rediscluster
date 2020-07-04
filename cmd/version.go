package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var version string

func SetVersion(v string) {
	version = v
}

type versionCmd struct {
	out io.Writer
}

func newVersionCmd(out io.Writer) *cobra.Command {
	version := &versionCmd{out}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "plugin version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("no arguments accepted for this command")
			}
			return version.run()
		},
	}
	return cmd
}

func (v *versionCmd) run() error {
	_, err := fmt.Fprintf(v.out, "Plugin Version: %s\n", version)
	if err != nil {
		return err
	}
	return nil
}
