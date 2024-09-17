package awsp

import (
	_ "embed"
	"os"
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
			MainCmd:     awspName,
			RootCmd:     config.RootCmd,
			PS1Cmd:      ps1Name,
			SourceStart: sourceStart,
		})
	},
}

func MainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: awspName,
	}

	cmd.AddCommand(ps1Cmd)
	cmd.AddCommand(useCmd)
	cmd.AddCommand(refreshCmd)
	cmd.AddCommand(processCmd)
	cmd.AddCommand(resetCmd)

	return cmd
}

type tmplInput struct {
	MainCmd     string
	RootCmd     string
	PS1Cmd      string
	SourceStart string
}

//go:embed awsp.sh
var awspBody string

const (
	awspName = "awsp"

	awsRegionEnvVar          = "AWS_DEFAULT_REGION"
	awsAccessKeyIDEnvVar     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"

	sourceStart = "### * aws-prompt awsp source start * ###"
)
