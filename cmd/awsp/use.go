package awsp

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/server"
)

func useCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:  useName,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			creds, err := server.GetCreds(args[0])
			if err != nil {
				return err
			}

			return dumpCreds(creds, stdout)
		},
	}
}

func refreshCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:  refreshName,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			accessKeyID := os.Getenv(awsAccessKeyIDEnvVar)
			if accessKeyID == "" {
				return nil
			}

			creds, err := server.Refresh(accessKeyID)
			if err != nil {
				return err
			}

			return dumpCreds(creds, stdout)
		},
	}
}

func dumpCreds(creds aws.Credentials, stdout io.Writer) error {
	tmpl, err := template.New(useName).Parse(useBody)
	if err != nil {
		return err
	}

	return tmpl.Execute(stdout, map[string]string{
		awsAccessKeyIDEnvVar:     fmt.Sprintf("%q", creds.AccessKeyID),
		awsSecretAccessKeyEnvVar: fmt.Sprintf("%q", creds.SecretAccessKey),
		awsSessionTokenEnvVar:    fmt.Sprintf("%q", creds.SessionToken),
	})
}

//go:embed use.sh
var useBody string

const (
	useName     = "use"
	refreshName = "refresh"

	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)
