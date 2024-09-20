package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/cmd/awsp"
	"kevwargo/aws-prompt/cmd/shellinit"
	"kevwargo/aws-prompt/internal/creds/cache"
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

	shellinit.InitCommand(rootCmd)

	rootCmd.AddCommand(awsp.MainCmd)
	rootCmd.AddCommand(cache.RunServerCmd)

	initBashCompletionCommand(rootCmd)

	return rootCmd.Execute()
}

func initBashCompletionCommand(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:       "bash-completion",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{awsp.MainCmd.Name()},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return rootCmd.GenBashCompletionV2(rootCmd.OutOrStdout(), false)
			} else if args[0] == awsp.MainCmd.Name() {
				return awsp.MainCmd.GenBashCompletionV2(awsp.MainCmd.OutOrStdout(), false)
			}

			return fmt.Errorf("cannot generate bash completion for standalone %q command", args[0])
		},
	}

	rootCmd.AddCommand(cmd)
}
