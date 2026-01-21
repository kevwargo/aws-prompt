package awsp

import (
	"github.com/spf13/cobra"
)

var MainCmd = createMainCommand()

func createMainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "awsp",
	}

	cmd.AddCommand(createUseCommand())
	cmd.AddCommand(createRefreshCommand())
	cmd.AddCommand(createResetCommand())
	cmd.AddCommand(createProcessCommand())

	return cmd
}
