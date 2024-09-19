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

func Command(rootCmdName string, awspCommands awsp.Commands) *cobra.Command {
	return &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := template.New(awspCommands.Main.Name()).Parse(shellBody)
			if err != nil {
				return err
			}

			var sourcableCommands []string
			for _, c := range awspCommands.Main.Commands() {
				sourcableCommands = append(sourcableCommands, fmt.Sprintf("%q", c.Name()))
				for _, a := range c.Aliases {
					sourcableCommands = append(sourcableCommands, fmt.Sprintf("%q", a))
				}
			}

			return tmpl.Execute(os.Stdout, tmplInput{
				MainCmd:            awspCommands.Main.Name(),
				RootCmd:            rootCmdName,
				PS1Cmd:             awspCommands.PS1.Name(),
				SourceStart:        awsp.SourceStart,
				SourcableCommands:  strings.Join(sourcableCommands, "|"),
				CompletionCommands: fmt.Sprintf("%q|%q", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd),
			})
		},
	}
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
