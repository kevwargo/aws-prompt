package profile

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/smithy-go"
)

func Resolve(name Name) (aws.Credentials, error) {
	ctx := context.Background()

	creds, err := load(ctx, name)
	if err != nil {
		if err = tryRelogin(err, name); err == nil {
			creds, err = load(ctx, name)
		}
	}
	if err != nil {
		return aws.Credentials{}, err
	}

	log.Printf("Loaded the config for %q", name)

	return creds, nil
}

func load(ctx context.Context, name Name) (aws.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(string(name)),
		config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
			o.TokenProvider = createMFAProvider(o, name)
		}),
	)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("loading config for profile %q: %w", name, err)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("retrieving creds for profile %q: %w", name, err)
	}

	return creds, nil
}

func tryRelogin(err error, profileName Name) error {
	var opErr *smithy.OperationError
	if !errors.As(err, &opErr) || opErr.Operation() != ssooidcCreateTokenOp || opErr.Service() != ssooidc.ServiceID {
		return err
	}

	log.Printf("SSO token refresh failed for %q, attempting re-login ...", profileName)

	cmd := exec.Command("aws", "sso", "login", "--profile", string(profileName))
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func createMFAProvider(o *stscreds.AssumeRoleOptions, profileName Name) func() (string, error) {
	promptParts := []string{fmt.Sprintf("profile:%q", profileName)}
	if o.RoleARN != "" {
		promptParts = append(promptParts, fmt.Sprintf("roleArn:%q", o.RoleARN))
	}
	if o.SerialNumber != nil {
		promptParts = append(promptParts, fmt.Sprintf("serial:%q", *o.SerialNumber))
	}

	prompt := fmt.Sprintf("Provide MFA one time code for (%s): ", strings.Join(promptParts, ", "))

	return func() (string, error) {
		_, err := fmt.Fprint(os.Stderr, prompt)
		if err != nil {
			return "", err
		}

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(line), nil
	}
}

const (
	ssooidcCreateTokenOp = "CreateToken"
)
