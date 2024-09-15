package awsp

import (
	_ "embed"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/creds"
)

func useCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     useName,
		Aliases: []string{"u"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			c, err := creds.Get(args[0])
			if err != nil {
				return err
			}

			return dumpCreds(c, stdout)
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

			c, err := creds.Refresh(accessKeyID)
			if err != nil {
				return err
			}

			return dumpCreds(c, stdout)
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

func processCommand(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:     processName,
		Aliases: []string{"p"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			c, err := creds.ResolveProcess(args[0], args[1:])
			if err != nil {
				return err
			}

			return dumpCreds(c, stdout)
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
		return emptyCredsMap
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
	processName = "process"
	resetName   = "reset"
)

var emptyCredsMap = map[string]string{
	awsAccessKeyIDEnvVar:     "",
	awsSecretAccessKeyEnvVar: "",
	awsSessionTokenEnvVar:    "",
}
