package awsp

import (
	"github.com/spf13/cobra"
)

type Commands struct {
	Main *cobra.Command
	PS1  *cobra.Command
}

func InitCommands() Commands {
	mainCmd := &cobra.Command{
		Use: "awsp",
	}

	mainCmd.AddCommand(ps1Cmd)
	mainCmd.AddCommand(useCmd)
	mainCmd.AddCommand(refreshCmd)
	mainCmd.AddCommand(processCmd)
	mainCmd.AddCommand(resetCmd)

	return Commands{
		Main: mainCmd,
		PS1:  ps1Cmd,
	}
}

const (
	SourceStart = "### * aws-prompt awsp source start * ###"

	awsRegionEnvVar          = "AWS_DEFAULT_REGION"
	awsAccessKeyIDEnvVar     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)
