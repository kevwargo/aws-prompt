package cache

import (
	"log"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/profile"
)

// Credentials cache, which automatically connects to the Unix socket
// of the credentials cache server.
// The zero value is ready for use.
type Cache struct {
	client *rpc.Client
}

type GetResp struct {
	Creds *aws.Credentials
}

type StoreRequest struct {
	Profile profile.Name
	Creds   aws.Credentials
}

// The default instance of cache for convenience.
var Default = &Cache{}

func (c *Cache) Get(name profile.Name) (*aws.Credentials, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	var resp GetResp
	if err := c.client.Call(serverName+".Get", name, &resp); err != nil {
		return nil, err
	}

	return resp.Creds, nil
}

func (c *Cache) Store(profileName profile.Name, creds aws.Credentials) error {
	if err := c.connect(); err != nil {
		return err
	}

	var resp struct{}

	return c.client.Call(serverName+".Store", StoreRequest{Profile: profileName, Creds: creds}, &resp)
}

func (c *Cache) Info(accessKeyID string) (info awskey.Info, err error) {
	if err = c.connect(); err != nil {
		return awskey.Info{}, err
	}

	err = c.client.Call(serverName+".Info", accessKeyID, &info)

	return
}

func (c *Cache) List() (profiles []string, err error) {
	if err = c.connect(); err != nil {
		return nil, err
	}

	err = c.client.Call(serverName+".List", struct{}{}, &profiles)

	return
}

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
