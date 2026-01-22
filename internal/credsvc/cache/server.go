package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/awskey"
	"kevwargo/aws-prompt/internal/credsvc/profile"
	"kevwargo/aws-prompt/internal/regionsvc"
)

var RunServerCmd = &cobra.Command{
	Use:    "_run_server",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		srv := server{
			profileCreds:    make(map[profile.Name]aws.Credentials),
			accessKeyInfo:   make(map[string]awskey.Info),
			accountRegions:  make(map[string][]regionsvc.Region),
			regionsResolver: regionsvc.NewResolver(),
		}

		return srv.run()
	},
}

type server struct {
	profileCreds      map[profile.Name]aws.Credentials
	profileCredsMutex sync.Mutex

	accountRegions      map[string][]regionsvc.Region
	accountRegionsMutex sync.Mutex
	regionsResolver     *regionsvc.Resolver

	accessKeyInfo      map[string]awskey.Info
	accessKeyInfoMutex sync.RWMutex
}

func (s *server) Get(name profile.Name, resp *GetResp) error {
	resp.Creds = s.getCreds(name)
	return nil
}

func (s *server) Store(req StoreRequest, resp *struct{}) error {
	if err := s.storeAccessKey(req); err != nil {
		return err
	}
	s.storeProfile(req.Profile, req.Creds)

	return nil
}

func (s *server) Info(accessKeyID string, resp *awskey.Info) (err error) {
	*resp, err = s.getKeyInfo(accessKeyID)

	return err
}

func (s *server) ListProfiles(req struct{}, resp *[]profile.Name) error {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()

	for p, creds := range s.profileCreds {
		if !creds.Expired() {
			*resp = append(*resp, p)
		} else {
			delete(s.profileCreds, p)
		}
	}

	return nil
}

func (s *server) ListRegions(accessKeyID string, resp *[]regionsvc.Region) error {
	s.accountRegionsMutex.Lock()
	defer s.accountRegionsMutex.Unlock()

	info, err := s.getKeyInfo(accessKeyID)
	if err != nil {
		return err
	}

	if regions, ok := s.accountRegions[info.AccountID]; ok {
		*resp = regions
		return nil
	}

	creds := s.getCreds(info.Profile)
	if creds == nil {
		// what?
		return nil
	}

	ctx := context.Background()
	awscfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(
		aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return *creds, nil
		}),
	))
	if err != nil {
		return err
	}

	regions, err := s.regionsResolver.List(ctx, awscfg)
	if err != nil {
		return err
	}

	s.accountRegions[info.AccountID] = regions
	*resp = regions

	return nil
}

func (s *server) getCreds(name profile.Name) *aws.Credentials {
	s.profileCredsMutex.Lock()
	defer s.profileCredsMutex.Unlock()

	creds, ok := s.profileCreds[name]
	if !ok {
		return nil
	}

	if !creds.Expired() {
		return &creds
	}

	log.Printf("Removing creds for %q expired on %s", name, creds.Expires)
	delete(s.profileCreds, name)

	return nil
}

func (s *server) getKeyInfo(accessKeyID string) (awskey.Info, error) {
	s.accessKeyInfoMutex.RLock()
	defer s.accessKeyInfoMutex.RUnlock()

	if info, ok := s.accessKeyInfo[accessKeyID]; ok {
		return info, nil
	}

	accountID, err := awskey.DecodeAccountID(accessKeyID)
	if err != nil {
		return awskey.Info{}, err
	}

	return awskey.Info{AccountID: accountID}, nil
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
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
