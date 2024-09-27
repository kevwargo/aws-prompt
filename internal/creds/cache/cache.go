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

	activeClient *client
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
}

func WithCache(fn func(Cache) error) (err error) {
	if activeClient == nil {
		activeClient, err = connect()
		if err != nil {
			return err
		}

		defer func() {
			if err := activeClient.Close(); err != nil {
				log.Printf("closing cache server client: %s", err.Error())
			}
			activeClient = nil
		}()
	}

	return fn(activeClient)
}

type GetResp struct {
	Creds *aws.Credentials
}

type StoreRequest struct {
	Profile profile.Name
	Creds   aws.Credentials
}
