package cmd

import (
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/cmd/awsp"
	"kevwargo/aws-prompt/cmd/shellinit"
	"kevwargo/aws-prompt/internal/credsvc/cache"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:           "aws-prompt",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	rootCmd.AddCommand(awsp.MainCmd)
	rootCmd.AddCommand(cache.RunServerCmd)
	shellinit.InitCommand(rootCmd)

	return rootCmd.Execute()
}
