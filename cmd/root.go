package cmd

import (
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/cmd/awsp"
	"kevwargo/aws-prompt/internal/config"
	"kevwargo/aws-prompt/internal/creds/cache"
)

func Execute() error {
	cmd := &cobra.Command{
		Use:           config.RootCmd,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(awsp.InitCmd, awsp.MainCommand())
	cmd.AddCommand(cache.RunServerCmd)

	return cmd.Execute()
}
