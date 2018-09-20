package commands

import (
	logger "log"
	"os"

	"github.com/spf13/cobra"
)

type rootFlags struct {
	Dir     string
	Verbose bool
}

var (
	rootCmdFlags rootFlags
	RootCmd      = &cobra.Command{
		Use:   "blogen",
		Short: "Static blog generator",
	}

	log = logger.New(os.Stderr, "", 0)
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&rootCmdFlags.Dir, "dir", "d", ".", "The source directory")
	RootCmd.PersistentFlags().BoolVarP(&rootCmdFlags.Verbose, "verbose", "v", false, "Verbose output")
}
