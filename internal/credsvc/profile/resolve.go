package profile

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
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

	creds, err := retrieveCredsWithRetry(ctx, cfg.Credentials, name)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("retrieving creds for profile %q: %w", name, err)
	}

	return creds, nil
}

func retrieveCredsWithRetry(ctx context.Context, provider aws.CredentialsProvider, profileName Name) (aws.Credentials, error) {
	creds, err := provider.Retrieve(ctx)

	if isErrSSOCreateToken(err) || isErrSSOCachedTokenFile(err) {
		cmd := exec.Command("aws", "sso", "login", "--profile", string(profileName))
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return aws.Credentials{}, err
		}

		return provider.Retrieve(ctx)
	}

	return creds, err
}

func isErrSSOCreateToken(err error) bool {
	if opErr, ok := errors.AsType[*smithy.OperationError](err); ok {
		if opErr.Operation() == ssooidcCreateTokenOp && opErr.Service() == ssooidc.ServiceID {
			log.Printf("Retrying after SSO token creation error: %s", err.Error())
			return true
		}
	}

	return false
}

func isErrSSOCachedTokenFile(err error) bool {
	if errors.Is(err, fs.ErrNotExist) {
		if strings.Contains(err.Error(), "failed to read cached SSO token file") {
			log.Printf("Retrying after failing to read cached SSO token file: %s", err.Error())
			return true
		}
	}

	return false
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
