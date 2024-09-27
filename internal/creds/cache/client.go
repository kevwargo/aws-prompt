package cache

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/profile"
)

type client struct {
	*rpc.Client
}

func (c *client) Get(name profile.Name) (*aws.Credentials, error) {
	var resp GetResp
	if err := c.Call(serverName+".Get", name, &resp); err != nil {
		return nil, err
	}

	return resp.Creds, nil
}

func (c *client) Store(req StoreRequest) error {
	var resp struct{}

	return c.Call(serverName+".Store", req, &resp)
}

func (c *client) Info(accessKeyID string) (info awskey.Info, err error) {
	err = c.Call(serverName+".Info", accessKeyID, &info)
	return
}

func (c *client) List() (profiles []string, err error) {
	err = c.Call(serverName+".List", struct{}{}, &profiles)
	return
}

func connect() (*client, error) {
	c, err := rpc.Dial("unix", socketPath)
	if err == nil {
		return &client{c}, nil
	}

	if err := handleConnectError(err); err != nil {
		return nil, err
	}
	if err := startServer(); err != nil {
		return nil, err
	}

	for start := time.Now(); time.Since(start) < 2*time.Second; time.Sleep(10 * time.Millisecond) {
		c, err = rpc.Dial("unix", socketPath)
		if err == nil {
			return &client{c}, nil
		}
	}

	return nil, err
}

func startServer() error {
	rootExec, err := os.Executable()
	if err != nil {
		rootExec = os.Args[0]
	}

	lf, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening log file %s: %w", logFile, err)
	}

	cmd := exec.Command(rootExec, RunServerCmd.Use)
	cmd.Stdout = lf
	cmd.Stderr = lf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %q: %w", cmd.String(), err)
	}

	return nil
}

func handleConnectError(connErr error) error {
	if errors.Is(connErr, syscall.ECONNREFUSED) {
		s, err := os.Stat(socketPath)
		if err != nil {
			return connErr
		}

		if s.Mode().Type()&os.ModeSocket == 0 {
			return fmt.Errorf("%w: %s is not a socket", connErr, socketPath)
		}

		if err := os.Remove(socketPath); err != nil {
			return connErr
		}

		return nil
	}

	if errors.Is(connErr, syscall.ENOENT) {
		return nil
	}

	return connErr
}
