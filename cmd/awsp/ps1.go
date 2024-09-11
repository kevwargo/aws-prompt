package awsp

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
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

	var label string
	expiration := "?"

	if status == nil {
		label, err = awskey.DecodeAccountID(accessKeyID)
		if err != nil {
			return "", err
		}
	} else {
		label = status.Profile
		if status.CanExpire {
			expiration = formatExpiration(time.Until(status.Expiration))
		}
	}

	if region := os.Getenv(awsRegionEnvVar); region != "" {
		label += ":" + shortenRegion(region)
	}

	return fmt.Sprintf("{%s (%s)}", colorize(label, colorPurple), expiration), nil
}

func formatExpiration(exp time.Duration) (text string) {
	minutes := exp.Minutes()
	color := colorGreen

	if minutes < 1 {
		color = colorBoldRed
		text = "exp"
	} else {
		text = strings.TrimSuffix(exp.Round(time.Minute).String(), "0s")

		if minutes < 10 {
			color = colorRed
		} else if minutes < 20 {
			color = colorYellow
		}
	}

	return colorize(text, color)
}

func colorize(text, color string) string {
	return fmt.Sprintf(`\[\e[%sm\]%s\[\e[0m\]`, color, text)
}

//go:embed ps1.sh
var ps1Body string

const (
	ps1Name = "ps1"

	awsAccessKeyIDEnvVar = "AWS_ACCESS_KEY_ID"
	awsRegionEnvVar      = "AWS_DEFAULT_REGION"

	colorPurple  = "38;5;56"
	colorGreen   = "32"
	colorYellow  = "33"
	colorRed     = "31"
	colorBoldRed = "1;31"
)
