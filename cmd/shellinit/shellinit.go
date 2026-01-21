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
	rootCmd.AddCommand(
		&cobra.Command{
			Use:  "init",
			RunE: runInit,
		},
		&cobra.Command{
			Use:       "bash-completion",
			Args:      cobra.MaximumNArgs(1),
			ValidArgs: []string{awsp.MainCmd.Name()},
			RunE:      runCompletion,
		},
		&cobra.Command{
			Use:    ps1Name,
			Hidden: true,
			RunE:   runPS1,
		},
	)

	if os.Getenv(preserveStdoutEnv) == "1" {
		rootCmd.SetOut(os.Stderr)
	}
}

type tmplInput struct {
	MainCmd            string
	RootCmd            string
	PS1Cmd             string
	SourcableCommands  string
	CompletionCommands string
	PreserveStdoutEnv  string
}

func runInit(_ *cobra.Command, _ []string) error {
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
		RootCmd:            os.Args[0],
		PS1Cmd:             ps1Name,
		SourcableCommands:  strings.Join(sourcableCommands, "|"),
		CompletionCommands: fmt.Sprintf("%q|%q", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd),
		PreserveStdoutEnv:  preserveStdoutEnv,
	})
}

func runCompletion(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Root().GenBashCompletionV2(os.Stdout, false)
	} else if args[0] == awsp.MainCmd.Name() {
		return awsp.MainCmd.GenBashCompletionV2(os.Stdout, false)
	}

	return fmt.Errorf("cannot generate bash completion for standalone %q command", args[0])
}

const preserveStdoutEnv = "AWSP_PRESERVE_STDOUT"

//go:embed init.sh
var shellBody string
