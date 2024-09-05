package cmd

import (
	"github.com/spf13/cobra"
)

func Execute() error {
	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           "aws-prompt",
	}

	return cmd.Execute()
}
