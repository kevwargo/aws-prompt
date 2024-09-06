package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Server struct {
	profileCreds     map[string]aws.Credentials
	accessKeyDetails map[string]AccessKeyDetails
	logFile          *os.File
	logger           *log.Logger
}

type AccessKeyDetails struct {
	AccessKeyID string
	Profile     string
	Expiration  time.Time
	CanExpire   bool
}

func newServer() (Server, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return Server{}, fmt.Errorf("locating use cache dir: %w", err)
	}

	logFileName := filepath.Join(cacheDir, "aws-prompt-server.log")
	logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return Server{}, fmt.Errorf("opening log file %q: %w", logFileName, err)
	}

	return Server{
		profileCreds:     make(map[string]aws.Credentials),
		accessKeyDetails: make(map[string]AccessKeyDetails),
		logFile:          logFile,
		logger:           log.New(logFile, "| ", log.LstdFlags|log.Lmsgprefix),
	}, nil
}

func (s Server) GetCreds(profile string, resp *aws.Credentials) error {
	return s.getCreds(profile, resp, true)
}

func (s Server) getCreds(profile string, resp *aws.Credentials, useCache bool) error {
	if useCache {
		if cached, ok := s.profileCreds[profile]; ok && !cached.Expired() {
			*resp = cached
			return nil
		}
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return fmt.Errorf("loading config for profile %q: %w", profile, err)
	}

	s.logger.Printf("Loaded the config for %q", profile)

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("retrieving creds for profile %q: %w", profile, err)
	}

	s.logger.Printf("Retrieved %s (expires on %s)", creds.AccessKeyID, creds.Expires)

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

func (s Server) Status(accessKeyID string, resp *AccessKeyDetails) error {
	*resp = s.accessKeyDetails[accessKeyID]
	return nil
}

func (s Server) Refresh(accessKeyID string, resp *aws.Credentials) error {
	details, ok := s.accessKeyDetails[accessKeyID]
	if !ok {
		return fmt.Errorf("access key %s not recognized", accessKeyID)
	}

	return s.getCreds(details.Profile, resp, false)
}
