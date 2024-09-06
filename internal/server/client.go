package server

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var serverName = reflect.TypeOf((*Server)(nil)).Elem().Name()

func GetCreds(profile string) (aws.Credentials, error) {
	client, err := connect()
	if err != nil {
		return aws.Credentials{}, err
	}

	var creds aws.Credentials
	if err := client.Call(serverName+".GetCreds", profile, &creds); err != nil {
		return aws.Credentials{}, err
	}

	return creds, nil
}

func Status(accessKeyID string) (AccessKeyDetails, error) {
	client, err := connect()
	if err != nil {
		return AccessKeyDetails{}, err
	}

	var details AccessKeyDetails
	if err := client.Call(serverName+".Status", accessKeyID, &details); err != nil {
		return AccessKeyDetails{}, err
	}

	return details, nil
}

func Refresh(accessKeyID string) (aws.Credentials, error) {
	client, err := connect()
	if err != nil {
		return aws.Credentials{}, err
	}

	var creds aws.Credentials
	if err := client.Call(serverName+".Refresh", accessKeyID, &creds); err != nil {
		return aws.Credentials{}, err
	}

	return creds, nil
}

func connect() (*rpc.Client, error) {
	client, err := rpc.Dial("unix", SocketPath)
	if err == nil {
		return client, nil
	}

	if err := handleConnectError(err); err != nil {
		return nil, err
	}

	if err := startServer(); err != nil {
		return nil, err
	}

	for start := time.Now(); time.Since(start) < 2*time.Second; time.Sleep(10 * time.Millisecond) {
		client, err = rpc.Dial("unix", SocketPath)
		if err == nil {
			break
		}
	}

	return client, err
}

func startServer() error {
	rootExec, err := os.Executable()
	if err != nil {
		rootExec = os.Args[0]
	}

	cmd := exec.Command(rootExec, RunCmd.Use)

	return cmd.Start()
}

func handleConnectError(connErr error) error {
	if errors.Is(connErr, syscall.ECONNREFUSED) {
		// The socket file exists but either not bound, or is not a socket
		s, err := os.Stat(SocketPath)
		if err != nil {
			return connErr
		}

		if s.Mode().Type()&os.ModeSocket == 0 {
			return fmt.Errorf("%w: %s is not a socket", connErr, SocketPath)
		}

		if err := os.Remove(SocketPath); err != nil {
			return connErr
		}

		return nil
	}

	if errors.Is(connErr, syscall.ENOENT) {
		return nil
	}

	return connErr
}
