package awsp

import (
	_ "embed"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/server"
)

func useCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     useName,
		Aliases: []string{"u"},
		Args:    cobra.ExactArgs(1),
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
		Use:     refreshName,
		Aliases: []string{"f"},
		Args:    cobra.NoArgs,
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

func resetCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     resetName,
		Aliases: []string{"x"},
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			for name := range mapCredEnvs(nil) {
				_, err := fmt.Fprintf(stdout, "unset %s;\n", name)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func dumpCreds(creds aws.Credentials, stdout io.Writer) error {
	for name, value := range mapCredEnvs(&creds) {
		_, err := fmt.Fprintf(stdout, "export %s=%q;\n", name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func mapCredEnvs(creds *aws.Credentials) map[string]string {
	if creds == nil {
		return map[string]string{
			awsAccessKeyIDEnvVar:     "",
			awsSecretAccessKeyEnvVar: "",
			awsSessionTokenEnvVar:    "",
		}
	}

	return map[string]string{
		awsAccessKeyIDEnvVar:     creds.AccessKeyID,
		awsSecretAccessKeyEnvVar: creds.SecretAccessKey,
		awsSessionTokenEnvVar:    creds.SessionToken,
	}
}

const (
	useName     = "use"
	refreshName = "refresh"
	resetName   = "reset"

	awsSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	awsSessionTokenEnvVar    = "AWS_SESSION_TOKEN"
)
