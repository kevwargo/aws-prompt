package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

const serverName = "awsp"

var RunCmd = &cobra.Command{
	Use:    "_run_server",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		return run()
	},
}

func run() error {
	srv := newServer()

	log.Printf("running server %d with pgrp %s", os.Getpid(), getPgrp())
	pgid, err := syscall.Setsid()
	log.Printf("setsid() = (%d, %v) pgrp:%s", pgid, err, getPgrp())

	if err := rpc.RegisterName(serverName, srv); err != nil {
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

func getPgrp() string {
	if pgid, err := syscall.Getpgid(0); err == nil {
		return strconv.Itoa(pgid)
	} else {
		return err.Error()
	}
}

func handleSignals(listener net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	s := <-c

	log.Printf("Caught the %q signal, closing server", s.String())

	listener.Close()
	os.Exit(0)
}
