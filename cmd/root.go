package cmd

import (
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/cmd/awsp"
	"kevwargo/aws-prompt/internal/config"
	"kevwargo/aws-prompt/internal/server"
)

func Execute() error {
	cmd := &cobra.Command{
		Use:           config.Name,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(awsp.InitCmd, awsp.MainCommand())
	cmd.AddCommand(server.RunCmd)

	return cmd.Execute()
}
