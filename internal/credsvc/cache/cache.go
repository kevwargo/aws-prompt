package cache

import (
	"bufio"
	"iter"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/credsvc/profile"
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

func (c *Cache) List() (profile.List, error) {
	if err := c.connect(); err != nil {
		return profile.List{}, err
	}

	var list profile.List
	if err := c.client.Call(serverName+".List", struct{}{}, &list.Active); err != nil {
		return profile.List{}, err
	}

	for _, f := range config.DefaultSharedConfigFiles {
		for p, err := range readProfiles(f) {
			if err != nil {
				return profile.List{}, err
			}

			if !slices.Contains(list.Active, p) {
				list.Inactive = append(list.Inactive, p)
			}
		}
	}

	list.Sort()

	return list, nil
}

func readProfiles(filename string) iter.Seq2[profile.Name, error] {
	return func(yield func(profile.Name, error) bool) {
		f, err := os.Open(filename)
		if err != nil {
			yield("", err)
			return
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			if m := regexConfigProfile.FindStringSubmatch(s.Text()); m != nil {
				if !yield(profile.Name(m[1]), nil) {
					return
				}
			}
		}
		if err := s.Err(); err != nil {
			yield("", err)
		}
	}
}

var (
	socketPath string
	logFile    string

	regexConfigProfile = regexp.MustCompile(`^\[profile +([^\]]+)\]`)
)

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("unable to locate user cache dir: %s", err.Error())
	}

	socketPath = filepath.Join(cacheDir, "aws-prompt-server.sock")
	logFile = filepath.Join(cacheDir, "aws-prompt-server.log")
}
