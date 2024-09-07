package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var SocketPath string

const serverName = "awsp"

var RunCmd = &cobra.Command{
	Use:    "_run_server",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return run()
	},
}

func run() error {
	srv, err := newServer()
	if err != nil {
		return err
	}

	if err := rpc.RegisterName(serverName, srv); err != nil {
		return fmt.Errorf("registering rpc: %w", err)
	}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", SocketPath, err)
	}
	go handleSignals(srv, listener)

	srv.logger.Printf("Accepting connections on %s ...", SocketPath)
	rpc.Accept(listener)

	return nil
}

func init() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("unable to locate user cache dir: %s", err.Error())
	}

	SocketPath = filepath.Join(cacheDir, "aws-prompt-server.sock")
}

func handleSignals(srv *Server, listener net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	s := <-c

	srv.logger.Printf("Caught the %s signal, closing server", s.String())
	srv.logFile.Close()

	listener.Close()

	os.Exit(0)
}
