package awsp

import (
	"github.com/spf13/cobra"
)

var MainCmd = createMainCommand()

func createMainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "awsp",
	}

	cmd.AddCommand(PS1Cmd)
	cmd.AddCommand(createUseCommand())
	cmd.AddCommand(createRefreshCommand())
	cmd.AddCommand(createResetCommand())
	cmd.AddCommand(createProcessCommand())

	return cmd
}

const (
	SourceStart = "### * aws-prompt awsp source start * ###"

	awsRegionEnvVar          = "AWS_DEFAULT_REGION"
	awsAccessKeyIDEnvVar     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)
