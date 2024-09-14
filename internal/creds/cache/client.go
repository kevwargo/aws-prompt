package cache

import (
	"errors"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"kevwargo/aws-prompt/internal/awskey"
)

type client struct {
	c *rpc.Client
}

func (c *client) Get(profile string) (*aws.Credentials, error) {
	var resp GetResp
	if err := c.c.Call(serverName+".Get", profile, &resp); err != nil {
		return nil, err
	}

	return resp.Creds, nil
}

func (c *client) Store(profile string, creds aws.Credentials) error {
	var (
		req  StoreRequest = StoreRequest{Profile: profile, Creds: creds}
		resp struct{}
	)

	return c.c.Call(serverName+".Store", req, &resp)
}

func (c *client) Info(accessKeyID string) (info awskey.Info, err error) {
	if err := c.c.Call(serverName+".Info", accessKeyID, &info); err != nil {
		return awskey.Info{}, err
	}

	return info, nil
}

func (c *client) Close() {
	if err := c.c.Close(); err != nil {
		log.Printf("closing cache server client: %s", err.Error())
	}
}

func (c *client) connect() (err error) {
	c.c, err = rpc.Dial("unix", socketPath)
	if err == nil {
		return nil
	}

	if err := handleConnectError(err); err != nil {
		return err
	}
	if err := startServer(); err != nil {
		return err
	}

	for start := time.Now(); time.Since(start) < 2*time.Second; time.Sleep(10 * time.Millisecond) {
		c.c, err = rpc.Dial("unix", socketPath)
		if err == nil {
			break
		}
	}

	return err
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
