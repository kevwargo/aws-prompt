package cache

import (
	"cmp"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/creds/profile"
)

var RunServerCmd = &cobra.Command{
	Use:    "_run_server",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		srv := server{
			profileCreds:  make(map[profile.Name]aws.Credentials),
			accessKeyInfo: make(map[string]awskey.Info),
		}

		return srv.run()
	},
}

type server struct {
	profileCreds      map[profile.Name]aws.Credentials
	profileCredsMutex sync.Mutex

	accessKeyInfo      map[string]awskey.Info
	accessKeyInfoMutex sync.Mutex
}

func (s *server) Get(name profile.Name, resp *GetResp) error {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()

	if creds, ok := s.profileCreds[name]; ok {
		if !creds.Expired() {
			resp.Creds = &creds
			return nil
		}

		log.Printf("Removing creds for %q expired on %s", name, creds.Expires)
		delete(s.profileCreds, name)
	}

	resp.Creds = nil
	return nil
}

func (s *server) Store(req StoreRequest, resp *struct{}) error {
	if err := s.storeAccessKey(req); err != nil {
		return err
	}
	s.storeProfile(req.Profile, req.Creds)

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

		*resp = awskey.Info{AccountID: accountID}
	}

	return nil
}

func (s *server) List(req struct{}, resp *[]profile.Name) error {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()

	for p, creds := range s.profileCreds {
		if !creds.Expired() {
			*resp = append(*resp, p)
		}
	}

	slices.SortFunc(*resp, func(a, b profile.Name) int {
		if a.IsPseudo() != b.IsPseudo() {
			// sort pseudo-profiles after the normal ones
			return -cmp.Compare(a, b)
		}
		return cmp.Compare(a, b)
	})

	return nil
}

func (s *server) storeProfile(profile profile.Name, creds aws.Credentials) {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()

	s.profileCreds[profile] = creds
}

func (s *server) storeAccessKey(req StoreRequest) error {
	s.accessKeyInfoMutex.Lock()
	defer s.accessKeyInfoMutex.Unlock()

	accountID, err := awskey.DecodeAccountID(req.Creds.AccessKeyID)
	if err != nil {
		return err
	}

	info := awskey.Info{
		AccountID: accountID,
		Profile:   req.Profile,
	}
	if req.Creds.CanExpire {
		info.Expiration = &req.Creds.Expires
	}
	s.accessKeyInfo[req.Creds.AccessKeyID] = info

	var expiration string
	if req.Creds.CanExpire {
		expiration = fmt.Sprintf(" (expiring on %s)", req.Creds.Expires)
	}

	log.Printf("Stored creds for profile %q%s", req.Profile, expiration)

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

	var caughtSignal os.Signal
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		caughtSignal = <-c
		listener.Close()
	}()

	log.Printf("Accepting connections on %s ...", socketPath)
	for {
		conn, err := listener.Accept()
		if err == nil {
			go rpc.DefaultServer.ServeConn(conn)
			continue
		}

		if errors.Is(err, net.ErrClosed) {
			log.Printf("Closing server due to signal %q", caughtSignal)
			return nil
		}

		return err
	}
}

const serverName = "awsp"
