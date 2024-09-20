package awsp

import (
	"github.com/spf13/cobra"
)

var MainCmd = &cobra.Command{
	Use: "awsp",
}

func init() {
	MainCmd.AddCommand(PS1Cmd)
	MainCmd.AddCommand(UseCmd)
	MainCmd.AddCommand(RefreshCmd)
	MainCmd.AddCommand(ProcessCmd)
	MainCmd.AddCommand(ResetCmd)
}

const (
	SourceStart = "### * aws-prompt awsp source start * ###"

	awsRegionEnvVar          = "AWS_DEFAULT_REGION"
	awsAccessKeyIDEnvVar     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)
