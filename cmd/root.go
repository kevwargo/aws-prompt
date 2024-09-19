package cmd

import (
	"fmt"

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
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	awspMainCmd := awsp.MainCommand()

	cmd.AddCommand(awsp.InitCmd, awspMainCmd)
	cmd.AddCommand(cache.RunServerCmd)

	initBashCompletionCommand(cmd, awspMainCmd)

	return cmd.Execute()
}

func initBashCompletionCommand(rootCmd, awspCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:       "bash-completion",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{awspCmd.Name()},
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return rootCmd.GenBashCompletionV2(rootCmd.OutOrStdout(), false)
			} else if args[0] == awspCmd.Name() {
				return awspCmd.GenBashCompletionV2(awspCmd.OutOrStdout(), false)
			}

			return fmt.Errorf("cannot generate bash completion for standalone %q command", args[0])
		},
	}

	rootCmd.AddCommand(cmd)
}
