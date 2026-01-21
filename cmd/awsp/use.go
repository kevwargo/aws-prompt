package awsp

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/credsvc"
	"kevwargo/aws-prompt/internal/credsvc/cache"
	"kevwargo/aws-prompt/internal/credsvc/profile"
)

func createUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:     useName,
		Aliases: []string{"u"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			creds, err := credsvc.Get(profile.Name(args[0]), false)
			if err != nil {
				return err
			}

			dumpCreds(creds)
			return nil
		},
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			names, err := generateCompletions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return names, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		},
	}
}

func createRefreshCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     refreshName,
		Aliases: []string{"f"},
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			accessKeyID := os.Getenv(credsvc.EnvAWSAccessKeyID)
			if accessKeyID == "" {
				return nil
			}

			creds, err := credsvc.Refresh(accessKeyID, force)
			if err != nil {
				return err
			}

			dumpCreds(creds)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force refresh the creds (don't use cached ones)")

	return cmd
}

func createResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     resetName,
		Aliases: []string{"x"},
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			for _, env := range credsvc.Envs {
				fmt.Printf("unset %s\n", env)
			}

			return nil
		},
	}
}

func createProcessCommand() *cobra.Command {
	return &cobra.Command{
		Use:     processName,
		Aliases: []string{"p"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			creds, err := credsvc.ResolveProcess(args[0], args[1:])
			if err != nil {
				return err
			}

			dumpCreds(creds)
			return nil
		},
	}
}

func dumpCreds(creds aws.Credentials) {
	for name, value := range credsvc.Map(creds) {
		fmt.Printf("export %s=%q\n", name, value)
	}
}

func generateCompletions() ([]string, error) {
	profiles, err := cache.Default.List()
	if err != nil {
		return nil, err
	}

	comps := make([]string, 0, len(profiles.Active)+len(profiles.Inactive))
	if len(profiles.Active) > 0 {
		for _, p := range profiles.Active {
			comps = append(comps, string(p))
		}
		comps = append(comps, "###")
	}
	for _, p := range profiles.Inactive {
		comps = append(comps, string(p))
	}

	return comps, nil
}

const (
	useName     = "use"
	refreshName = "refresh"
	processName = "process"
	resetName   = "reset"
)
