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

	awspCommands := awsp.InitCommands()
	shellInitCmd := shellinit.Command(rootCmd.Name(), awspCommands)

	rootCmd.AddCommand(shellInitCmd)
	rootCmd.AddCommand(awspCommands.Main)
	rootCmd.AddCommand(cache.RunServerCmd)

	initBashCompletionCommand(rootCmd, awspCommands.Main)

	return rootCmd.Execute()
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
