package cache

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/awskey"
)

var RunServerCmd = &cobra.Command{
	Use:    "_run_server",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		srv := server{
			profileCreds:  make(map[string]aws.Credentials),
			accessKeyInfo: make(map[string]awskey.Info),
		}

		return srv.run()
	},
}

type server struct {
	profileCreds      map[string]aws.Credentials
	profileCredsMutex sync.Mutex

	accessKeyInfo      map[string]awskey.Info
	accessKeyInfoMutex sync.Mutex
}

func (s *server) Get(profile string, resp *GetResp) error {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()

	if creds, ok := s.profileCreds[profile]; ok {
		if !creds.Expired() {
			resp.Creds = &creds
			return nil
		}

		delete(s.profileCreds, profile)
	}

	resp.Creds = nil
	return nil
}

func (s *server) Store(req StoreRequest, resp *struct{}) error {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()
	s.accessKeyInfoMutex.Lock()
	defer s.accessKeyInfoMutex.Unlock()

	s.profileCreds[req.Profile] = req.Creds

	keyInfo := awskey.Info{Profile: req.Profile}
	if req.Creds.CanExpire {
		keyInfo.Expiration = &req.Creds.Expires
	}
	s.accessKeyInfo[req.Creds.AccessKeyID] = keyInfo

	return nil
}

func (s *server) Info(accessKeyID string, resp *awskey.Info) error {
	s.accessKeyInfoMutex.Lock()
	defer s.accessKeyInfoMutex.Unlock()

	if info, ok := s.accessKeyInfo[accessKeyID]; ok {
		*resp = info
	} else {
		accountID, err := awskey.DecodeAccountID(accessKeyID)
		if err != nil {
			return err
		}

		*resp = awskey.Info{Profile: accountID}
	}

	return nil
}

func (s *server) run() error {
	if _, err := syscall.Setsid(); err != nil {
		return fmt.Errorf("calling setsid(): %w", err)
	}

	if err := rpc.RegisterName(serverName, s); err != nil {
		return fmt.Errorf("registering rpc: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", socketPath, err)
	}
	go handleSignals(listener)

	log.Printf("Accepting connections on %s ...", socketPath)
	rpc.Accept(listener)

	return nil
}

func handleSignals(listener net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	s := <-c

	log.Printf("Caught the %q signal, closing server", s.String())
	listener.Close()
}

const serverName = "awsp"
