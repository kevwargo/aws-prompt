package shellinit

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/cmd/awsp"
)

func InitCommand(rootCmd *cobra.Command) {
	initCmd := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := template.New(awsp.MainCmd.Name()).Parse(shellBody)
			if err != nil {
				return err
			}

			var sourcableCommands []string
			for _, c := range awsp.MainCmd.Commands() {
				sourcableCommands = append(sourcableCommands, fmt.Sprintf("%q", c.Name()))
				for _, a := range c.Aliases {
					sourcableCommands = append(sourcableCommands, fmt.Sprintf("%q", a))
				}
			}

			return tmpl.Execute(os.Stdout, tmplInput{
				MainCmd:            awsp.MainCmd.Name(),
				RootCmd:            rootCmd.Name(),
				PS1Cmd:             awsp.PS1Cmd.Name(),
				SourceStart:        awsp.SourceStart,
				SourcableCommands:  strings.Join(sourcableCommands, "|"),
				CompletionCommands: fmt.Sprintf("%q|%q", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd),
			})
		},
	}

	rootCmd.AddCommand(initCmd)
}

type tmplInput struct {
	MainCmd            string
	RootCmd            string
	PS1Cmd             string
	SourceStart        string
	SourcableCommands  string
	CompletionCommands string
}

//go:embed init.sh
var shellBody string
