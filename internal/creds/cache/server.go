package cache

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"

	"kevwargo/aws-prompt/internal/awskey"
)

var RunServerCmd = &cobra.Command{
	Use:    "_run_server",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		srv := server{
			profileCreds:  make(map[string]aws.Credentials),
			accessKeyInfo: make(map[string]*awskey.Info),
		}

		return srv.run()
	},
}

type server struct {
	profileCreds      map[string]aws.Credentials
	profileCredsMutex sync.Mutex

	accessKeyInfo      map[string]*awskey.Info
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

		log.Printf("Removing creds for %q expired on %s", profile, creds.Expires)
		delete(s.profileCreds, profile)
	}

	resp.Creds = nil
	return nil
}

func (s *server) Store(req StoreRequest, resp *struct{}) error {
	if err := s.storeAccessKey(req); err != nil {
		return err
	}
	if req.Profile != nil {
		s.storeProfile(*req.Profile, req.Creds)
	}

	return nil
}

func (s *server) storeProfile(profile string, creds aws.Credentials) {
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

	info := &awskey.Info{
		AccountID: accountID,
		Profile:   req.Profile,
	}
	if req.Creds.CanExpire {
		info.Expiration = &req.Creds.Expires
	}
	s.accessKeyInfo[req.Creds.AccessKeyID] = info

	identity := fmt.Sprintf("account %s", accountID)
	if req.Profile != nil {
		identity = fmt.Sprintf("profile %q", *req.Profile)
	}

	var expiration string
	if req.Creds.CanExpire {
		expiration = fmt.Sprintf(" (expiring on %s)", req.Creds.Expires)
	}

	log.Printf("Stored creds for %s%s", identity, expiration)

	go s.storeAssumedRole(req)

	return nil
}

func (s *server) storeAssumedRole(req StoreRequest) {
	credsFn := func(_ context.Context) (aws.Credentials, error) {
		return req.Creds, nil
	}
	stsOpts := sts.Options{
		Credentials: aws.CredentialsProviderFunc(credsFn),
		Region:      req.Region,
	}
	ctx := context.Background()

	resp, err := sts.New(stsOpts).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Printf("getting caller identity for %s: %s", req.Creds.AccessKeyID, err.Error())
		return
	}

	m := assumedRoleRegex.FindStringSubmatch(*resp.Arn)
	if m == nil {
		return
	}

	s.accessKeyInfoMutex.Lock()
	defer s.accessKeyInfoMutex.Unlock()

	if info := s.accessKeyInfo[req.Creds.AccessKeyID]; info != nil {
		assumedRole := fmt.Sprintf("%s/%s", m[1], m[2])
		info.AssumedRole = &assumedRole

		var profile string
		if req.Profile != nil {
			profile = fmt.Sprintf(" for profile %q", *req.Profile)
		}

		log.Printf("Stored AssumedRole %q%s", assumedRole, profile)
	}
}

func (s *server) Info(accessKeyID string, resp *awskey.Info) error {
	s.accessKeyInfoMutex.Lock()
	defer s.accessKeyInfoMutex.Unlock()

	if info := s.accessKeyInfo[accessKeyID]; info != nil {
		*resp = *info
	} else {
		accountID, err := awskey.DecodeAccountID(accessKeyID)
		if err != nil {
			return err
		}

		*resp = awskey.Info{AccountID: accountID}
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

var assumedRoleRegex = regexp.MustCompile("arn:aws:sts::([0-9]{12}):assumed-role/([^/]+)(/.*)?")
