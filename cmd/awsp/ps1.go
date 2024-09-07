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
	if status == nil {
		return awskey.DecodeAccountID(accessKeyID)
	}

	var expiration string
	if status.CanExpire {
		expiration = formatExpiration(status.Expiration)
	} else {
		expiration = "?"
	}

	return fmt.Sprintf("{%s%s%s (%s)}",
		colorPurple, status.Profile, colorEnd, expiration,
	), nil
}

func formatExpiration(exp time.Time) (text string) {
	minutes := time.Until(exp).Minutes()
	color := colorGreen

	if minutes < 1 {
		color = colorBoldRed
		text = "exp"
	} else {
		text = fmt.Sprintf("%dm", int(minutes))

		if minutes < 10 {
			color = colorRed
		} else if minutes < 20 {
			color = colorYellow
		}
	}

	return color + text + colorEnd
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
