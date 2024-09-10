package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/smithy-go"

	"kevwargo/aws-prompt/internal/awskey"
)

var (
	socketPath string
	logFile    string
)

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("unable to locate user cache dir: %s", err.Error())
	}

	socketPath = filepath.Join(cacheDir, "aws-prompt-server.sock")
	logFile = filepath.Join(cacheDir, "aws-prompt-server.log")
}

type Server struct {
	profileCreds   map[string]aws.Credentials
	profileCredsMu sync.Mutex

	accessKeyDetails   map[string]AccessKeyDetails
	accessKeyDetailsMu sync.Mutex
}

type AccessKeyDetails struct {
	AccessKeyID string
	Profile     string
	Expiration  time.Time
	CanExpire   bool
}

func newServer() *Server {
	return &Server{
		profileCreds:     make(map[string]aws.Credentials),
		accessKeyDetails: make(map[string]AccessKeyDetails),
	}
}

func (s *Server) GetCreds(profile string, resp *aws.Credentials) error {
	return s.getCreds(profile, resp, true)
}

func (s *Server) getCreds(profile string, resp *aws.Credentials, useCache bool) error {
	if useCache {
		if cached := s.getCachedCreds(profile); cached != nil {
			*resp = *cached
			return nil
		}
	}

	ctx := context.Background()
	creds, err := s.loadProfileCreds(ctx, profile)
	if err != nil {
		creds, err = s.tryRelogin(ctx, err, profile)
		if err != nil {
			return err
		}
	}

	accountID, err := awskey.DecodeAccountID(creds.AccessKeyID)
	if err != nil {
		accountID = fmt.Sprintf("%s(decodeErr:%s)", creds.AccessKeyID, err.Error())
	}

	log.Printf("Retrieved creds for %s (expire on %s)", accountID, creds.Expires)

	s.profileCredsMu.Lock()
	defer s.profileCredsMu.Unlock()
	s.accessKeyDetailsMu.Lock()
	defer s.accessKeyDetailsMu.Unlock()

	s.profileCreds[profile] = creds
	s.accessKeyDetails[creds.AccessKeyID] = AccessKeyDetails{
		AccessKeyID: creds.AccessKeyID,
		Profile:     profile,
		Expiration:  creds.Expires,
		CanExpire:   creds.CanExpire,
	}

	*resp = creds
	return nil
}

func (s *Server) loadProfileCreds(ctx context.Context, profile string) (aws.Credentials, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
			o.TokenProvider = createMFAProvider(o, profile)
		}),
	)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("loading config for profile %q: %w", profile, err)
	}

	log.Printf("Loaded the config for %q", profile)

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("retrieving creds for profile %q: %w", profile, err)
	}

	return creds, nil
}

func createMFAProvider(o *stscreds.AssumeRoleOptions, profile string) func() (string, error) {
	var createPromptCmd func(string) *exec.Cmd

	if _, err := exec.LookPath(kdialogCmd); err == nil {
		createPromptCmd = createKDialogPrompt
	} else if _, err := exec.LookPath(ksshaskpassCmd); err == nil {
		createPromptCmd = createKSSHAskPassPrompt
	} else {
		return func() (string, error) {
			return "", fmt.Errorf("Cannot interactively ask for MFA code, since neither %q nor %q are installed", kdialogCmd, ksshaskpassCmd)
		}
	}

	promptParts := []string{fmt.Sprintf("profile:%s", profile)}
	if o.RoleARN != "" {
		promptParts = append(promptParts, fmt.Sprintf("roleArn:%s", o.RoleARN))
	}
	if o.SerialNumber != nil {
		promptParts = append(promptParts, fmt.Sprintf("serial:%s", *o.SerialNumber))
	}

	cmd := createPromptCmd(fmt.Sprintf("Provide MFA one time code for (%s): ", strings.Join(promptParts, ", ")))

	return func() (string, error) {
		resp, err := cmd.Output()
		if err != nil {
			err = fmt.Errorf("executing %q: %w", cmd.String(), err)
			if ee := (*exec.ExitError)(nil); errors.As(err, &ee) && len(ee.Stderr) > 0 {
				err = fmt.Errorf("%w: %s", err, string(ee.Stderr))
			}

			return "", err
		}

		return strings.TrimSpace(string(resp)), nil
	}
}

func createKDialogPrompt(prompt string) *exec.Cmd {
	return exec.Command(kdialogCmd, "--title", "MFA", "--inputbox", prompt)
}

func createKSSHAskPassPrompt(prompt string) *exec.Cmd {
	return exec.Command(ksshaskpassCmd, prompt)
}

func (s *Server) tryRelogin(ctx context.Context, err error, profile string) (aws.Credentials, error) {
	var opErr *smithy.OperationError
	if !errors.As(err, &opErr) || opErr.Operation() != ssooidcCreateTokenOp || opErr.Service() != ssooidc.ServiceID {
		return aws.Credentials{}, err
	}

	log.Printf("SSO token refresh failed for %q, attempting re-login ...", profile)

	cmd := exec.Command("aws", "sso", "login", "--profile", profile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return aws.Credentials{}, err
	}

	return s.loadProfileCreds(ctx, profile)
}

func (s *Server) getCachedCreds(profile string) *aws.Credentials {
	s.profileCredsMu.Lock()
	defer s.profileCredsMu.Unlock()

	if cached, exists := s.profileCreds[profile]; exists {
		if !cached.Expired() {
			return &cached
		}
		delete(s.profileCreds, profile)
	}

	return nil
}

func (s *Server) Status(accessKeyID string, resp *AccessKeyDetails) error {
	if details := s.getKeyDetails(accessKeyID); details != nil {
		*resp = *details
	}

	return nil
}

func (s *Server) Refresh(accessKeyID string, resp *aws.Credentials) error {
	details := s.getKeyDetails(accessKeyID)
	if details == nil {
		return fmt.Errorf("access key %s not recognized", accessKeyID)
	}

	return s.getCreds(details.Profile, resp, false)
}

func (s *Server) getKeyDetails(accessKeyID string) *AccessKeyDetails {
	s.accessKeyDetailsMu.Lock()
	defer s.accessKeyDetailsMu.Unlock()

	if details, ok := s.accessKeyDetails[accessKeyID]; ok {
		return &details
	}

	return nil
}

const (
	ssooidcCreateTokenOp = "CreateToken"

	kdialogCmd     = "kdialog"
	ksshaskpassCmd = "ksshaskpass"
)
