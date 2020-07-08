package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var version string
var git string
var goversion string

func SetVersion(v string, g string, gv string) {
	version = v
	git = g
	goversion = gv
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
	_, err := fmt.Fprintf(v.out, "Plugin Version:\t%s\nGit:\t\t%s\nGo:\t\t%s\n",
		version, git, goversion)
	if err != nil {
		return err
	}
	return nil
}
