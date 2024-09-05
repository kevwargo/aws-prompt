package awsp

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"kevwargo/aws-prompt/internal/config"
)

func InitCommand() *cobra.Command {
	return &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := template.New(name).Parse(awspBody)
			if err != nil {
				return err
			}

			return tmpl.Execute(os.Stdout, tmplInput{
				Cmd:     name,
				RootCmd: config.Name,
			})
		},
	}
}

func MainCommand() *cobra.Command {
	cmds := []*cobra.Command{
		{
			Use: "ps1",
			Run: func(cmd *cobra.Command, args []string) { fmt.Println("echo ps1") },
		},
		{
			Use:     "use",
			Aliases: []string{"u"},
			Run:     func(cmd *cobra.Command, args []string) { fmt.Println("echo use") },
		},
		{
			Use:     "region",
			Aliases: []string{"r"},
			Run:     func(cmd *cobra.Command, args []string) { fmt.Println("echo region") },
		},
		{
			Use: "refresh",
			Run: func(cmd *cobra.Command, args []string) { fmt.Println("echo refresh") },
		},
	}

	cmd := &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("invalid command %s", strings.Join(args, " "))
		},
	}

	// prevent cobra from spewing garbage on stdout which would then be interpreted by bash
	cmd.SetOut(os.Stderr)

	cmd.AddCommand(cmds...)

	return cmd
}

type tmplInput struct {
	Cmd     string
	RootCmd string
}

const name = "awsp"

//go:embed awsp.sh
var awspBody string
