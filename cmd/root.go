package cmd

import (
	"github.com/spf13/cobra"
	"kevwargo/aws-prompt/cmd/awsp"
	"kevwargo/aws-prompt/internal/config"
)

func Execute() error {
	cmd := &cobra.Command{
		Use:           config.Name,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(awsp.InitCommand(), awsp.MainCommand())

	return cmd.Execute()
}
