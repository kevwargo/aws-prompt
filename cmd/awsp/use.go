package awsp

import (
	"fmt"
	"io"

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

			fmt.Fprintln(stdout, creds.AccessKeyID, creds.Expires)
			return nil
		},
	}
}

const useName = "use"
