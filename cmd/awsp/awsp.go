package awsp

import (
	"github.com/spf13/cobra"
)

var MainCmd = createMainCommand()

func createMainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "awsp",
	}

	cmd.AddCommand(
		createUseCommand(),
		createRefreshCommand(),
		createResetCommand(),
		createProcessCommand(),
		createRegionCommand(),
	)

	return cmd
}
