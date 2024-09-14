package cache

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"

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

type Cache interface {
	Get(profile string) (*aws.Credentials, error)
	Store(profile string, creds aws.Credentials) error
	Info(accessKeyID string) (awskey.Info, error)
	Close()
}

func Open() (Cache, error) {
	var c client
	if err := c.connect(); err != nil {
		return nil, err
	}

	return &c, nil
}

type GetResp struct {
	Creds *aws.Credentials
}

type StoreRequest struct {
	Profile string
	Creds   aws.Credentials
}
