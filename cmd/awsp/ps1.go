package awsp

import (
	_ "embed"
	"io"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"kevwargo/aws-prompt/internal/awskey"
)

func ps1Command(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use: ps1Name,
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := template.New(ps1Name).Parse(ps1Body)
			if err != nil {
				return err
			}

			keyID := os.Getenv("AWS_ACCESS_KEY_ID")

			var accountID string
			if keyID != "" {
				accountID, err = awskey.DecodeAccountID(keyID)
				if err != nil {
					return err
				}
			} else {
				accountID = "(empty)"
			}

			return tmpl.Execute(stdout, accountID)
		},
	}
}

//go:embed ps1.sh
var ps1Body string

const ps1Name = "ps1"
