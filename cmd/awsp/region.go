package awsp

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/credsvc"
	"kevwargo/aws-prompt/internal/credsvc/cache"
	"kevwargo/aws-prompt/internal/regionsvc"
)

func createRegionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "region",
		Aliases: []string{"r"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("export %s=%q\n", credsvc.EnvAWSRegion, args[0])
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			comps, err := generateRegionCompletions()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return comps, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
		},
	}
}

func generateRegionCompletions() (comps []cobra.Completion, err error) {
	var regions []regionsvc.Region
	if accessKeyID := os.Getenv(credsvc.EnvAWSAccessKeyID); accessKeyID != "" {
		regions, err = cache.Default.ListRegions(accessKeyID)
		if err != nil {
			return nil, err
		}
	}

	if len(regions) == 0 {
		regions = regionsvc.Defaults
	}

	for _, r := range regions {
		comps = append(comps, fmt.Sprintf("%s\t%s", r.Name, r.LongName))
	}

	return comps, nil
}
