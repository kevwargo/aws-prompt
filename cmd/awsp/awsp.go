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

var InitCmd = &cobra.Command{
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

func MainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           awspName,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("invalid command %s", strings.Join(args, " "))
		},
	}

	// we need to prevent any unwanted output from being written into stdout
	// because stdout is interpreted (sourced) as-is by bash
	cmd.SetOut(os.Stderr)
	cmd.PersistentPreRun = func(_ *cobra.Command, _ []string) { os.Stdout = os.Stderr }

	cmd.AddCommand(ps1Command(os.Stdout))
	cmd.AddCommand(useCommand(os.Stdout))
	cmd.AddCommand(refreshCommand(os.Stdout))
	cmd.AddCommand(processCommand(os.Stdout))
	cmd.AddCommand(resetCommand(os.Stdout))

	return cmd
}

type tmplInput struct {
	MainCmd string
	RootCmd string
	PS1Cmd  string
}

//go:embed awsp.sh
var awspBody string

const (
	awspName = "awsp"

	awsRegionEnvVar          = "AWS_DEFAULT_REGION"
	awsAccessKeyIDEnvVar     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)
