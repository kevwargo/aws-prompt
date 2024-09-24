package cache

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/profile"
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
	Get(name profile.Name) (*aws.Credentials, error)
	Store(req StoreRequest) error
	Info(accessKeyID string) (awskey.Info, error)
	List() ([]string, error)
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
	Profile profile.Name
	Creds   aws.Credentials
}
