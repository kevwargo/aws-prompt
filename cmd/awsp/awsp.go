package awsp

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"kevwargo/aws-prompt/internal/config"
)

func InitCommand() *cobra.Command {
	return &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := template.New(awspName).Parse(awspBody)
			if err != nil {
				return err
			}

			return tmpl.Execute(os.Stdout, tmplInput{
				MainCmd: awspName,
				RootCmd: config.Name,
				PS1Cmd:  ps1Name,
			})
		},
	}
}

func MainCommand() *cobra.Command {
	stdout := os.Stdout

	cmd := &cobra.Command{
		Use:           awspName,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// prevent all non-explicit writes to stdout
			os.Stdout = os.Stderr
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("invalid command %s", strings.Join(args, " "))
		},
	}

	// prevent cobra from spewing garbage on stdout which would then be interpreted by bash
	cmd.SetOut(os.Stderr)

	cmd.AddCommand(ps1Command(stdout))

	return cmd
}

type tmplInput struct {
	MainCmd string
	RootCmd string
	PS1Cmd  string
}

//go:embed awsp.sh
var awspBody string

const awspName = "awsp"
