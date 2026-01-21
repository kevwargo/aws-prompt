package credsvc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/processcreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"kevwargo/aws-prompt/internal/credsvc/cache"
	"kevwargo/aws-prompt/internal/credsvc/profile"
)

func ResolveProcess(name string, args []string) (aws.Credentials, error) {
	ctx := context.Background()

	cfg, err := loadProcessConfig(ctx, name, args)
	if err != nil {
		return aws.Credentials{}, err
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}

	profile, err := getPseudoProfile(ctx, cfg)
	if err != nil {
		return aws.Credentials{}, err
	}

	if err := cache.Default.Store(profile, creds); err != nil {
		return aws.Credentials{}, err
	}

	return creds, nil
}

func loadProcessConfig(ctx context.Context, name string, args []string) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(
			processcreds.NewProviderCommand(
				processcreds.NewCommandBuilderFunc(func(ctx context.Context) (*exec.Cmd, error) {
					cmd := exec.CommandContext(ctx, name, args...)
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					return cmd, nil
				}),
			),
		),
	)
}

func getPseudoProfile(ctx context.Context, cfg aws.Config) (profile.Name, error) {
	resp, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	m := assumedRoleRegex.FindStringSubmatch(*resp.Arn)
	if m == nil {
		return "", fmt.Errorf("invalid assumed role ARN %q", *resp.Arn)
	}

	return profile.Pseudo(m[1], m[2]), nil
}

var assumedRoleRegex = regexp.MustCompile("arn:aws:sts::([0-9]{12}):assumed-role/([^/]+)(/.*)?")
