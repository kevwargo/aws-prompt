package awsp

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/server"
)

func ps1Command(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use: ps1Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := template.New(ps1Name).Parse(ps1Body)
			if err != nil {
				return err
			}

			accessKeyID := os.Getenv(awsAccessKeyIDEnvVar)
			if accessKeyID == "" {
				return nil
			}

			data, err := describeAccessKey(accessKeyID)
			if err != nil {
				return err
			}

			return tmpl.Execute(stdout, data)
		},
	}
}

func describeAccessKey(accessKeyID string) (string, error) {
	status, err := server.Status(accessKeyID)
	if err != nil {
		return "", err
	}
	if status == (server.AccessKeyDetails{}) {
		return awskey.DecodeAccountID(accessKeyID)
	}

	expiresInMinutes := time.Until(status.Expires).Minutes()

	expiry, expiryColor := fmt.Sprintf("%dm", int(expiresInMinutes)), colorGreen
	if expiresInMinutes < 1 {
		expiry = "exp"
		expiryColor = colorBoldRed
	} else if expiresInMinutes < 10 {
		expiryColor = colorRed
	} else if expiresInMinutes < 25 {
		expiryColor = colorYellow
	}

	return fmt.Sprintf("{%s%s%s (%s%s%s)}",
		colorPurple, status.Profile, colorEnd, expiryColor, expiry, colorEnd,
	), nil
}

//go:embed ps1.sh
var ps1Body string

const (
	ps1Name = "ps1"

	awsAccessKeyIDEnvVar = "AWS_ACCESS_KEY_ID"

	colorPurple  = `\e[38;5;56m`
	colorGreen   = `\e[32m`
	colorYellow  = `\e[33m`
	colorRed     = `\e[31m`
	colorBoldRed = `\e[1;31m`
	colorEnd     = `\e[0m`
)
