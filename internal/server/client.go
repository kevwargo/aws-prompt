package server

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
)

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

func Status(accessKeyID string) (*AccessKeyDetails, error) {
	client, err := connect()
	if err != nil {
		return nil, err
	}

	var details AccessKeyDetails
	if err := client.Call(serverName+".Status", accessKeyID, &details); err != nil {
		return nil, err
	}

	if details.AccessKeyID != accessKeyID {
		return nil, nil
	}

	return &details, nil
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
	client, err := rpc.Dial("unix", socketPath)
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
		client, err = rpc.Dial("unix", socketPath)
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

	lf, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening log file %s: %w", logFile, err)
	}

	cmd := exec.Command(rootExec, RunCmd.Use)
	cmd.Stdout = lf
	cmd.Stderr = lf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %q: %w", cmd.String(), err)
	}

	log.Printf("started server %d in the background (caller pid:%d pgrp:%s)", cmd.Process.Pid, os.Getpid(), getPgrp())

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
