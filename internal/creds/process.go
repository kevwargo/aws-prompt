package creds

import (
	"context"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/processcreds"

	"kevwargo/aws-prompt/internal/creds/cache"
)

func ResolveProcess(name string, args []string) (aws.Credentials, error) {
	ctx := context.Background()

	creds, err := processcreds.NewProviderCommand(processcreds.NewCommandBuilderFunc(func(ctx context.Context) (*exec.Cmd, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		return cmd, nil
	})).Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Credentials{}, err
	}

	c, err := cache.Open()
	if err != nil {
		return aws.Credentials{}, err
	}
	defer c.Close()

	if err := c.Store(cache.StoreRequest{
		Creds:  creds,
		Region: cfg.Region,
	}); err != nil {
		return aws.Credentials{}, err
	}

	return creds, nil
}
