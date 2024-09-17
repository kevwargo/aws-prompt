package awsp

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/creds"
)

var useCmd = &cobra.Command{
	Use:     useName,
	Aliases: []string{"u"},
	Args:    cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		c, err := creds.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Println(sourceStart)
		dumpCreds(c)
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeProfiles(), cobra.ShellCompDirectiveNoFileComp
	},
}

var refreshCmd = &cobra.Command{
	Use:     refreshName,
	Aliases: []string{"f"},
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		accessKeyID := os.Getenv(awsAccessKeyIDEnvVar)
		if accessKeyID == "" {
			return nil
		}

		c, err := creds.Refresh(accessKeyID)
		if err != nil {
			return err
		}

		fmt.Println(sourceStart)
		dumpCreds(c)
		return nil
	},
}

var resetCmd = &cobra.Command{
	Use:     resetName,
	Aliases: []string{"x"},
	Args:    cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(sourceStart)

		for _, env := range []string{awsAccessKeyIDEnvVar, awsSecretAccessKeyEnvVar, awsSessionTokenEnvVar} {
			fmt.Printf("unset %s\n", env)
		}

		return nil
	},
}

var processCmd = &cobra.Command{
	Use:     processName,
	Aliases: []string{"p"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		c, err := creds.ResolveProcess(args[0], args[1:])
		if err != nil {
			return err
		}

		dumpCreds(c)
		return nil
	},
}

func dumpCreds(creds aws.Credentials) {
	for name, value := range mapCredEnvs(creds) {
		fmt.Printf("export %s=%q\n", name, value)
	}
}

func mapCredEnvs(creds aws.Credentials) map[string]string {
	return map[string]string{
		awsAccessKeyIDEnvVar:     creds.AccessKeyID,
		awsSecretAccessKeyEnvVar: creds.SecretAccessKey,
		awsSessionTokenEnvVar:    creds.SessionToken,
	}
}

var regexConfigProfile = regexp.MustCompile(`^\[profile +([^\]]+)\]`)

func completeProfiles() (profiles []string) {
	profileNames := make(map[string]struct{})

	for _, f := range config.DefaultSharedConfigFiles {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(b), "\n") {
			m := regexConfigProfile.FindStringSubmatch(line)
			if len(m) > 1 {
				profileNames[m[1]] = struct{}{}
			}
		}
	}

	for p := range profileNames {
		profiles = append(profiles, p)
	}

	return
}

const (
	useName     = "use"
	refreshName = "refresh"
	processName = "process"
	resetName   = "reset"
)
