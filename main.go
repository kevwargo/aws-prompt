package main

import (
	"fmt"
	"log"
	"os"

	"kevwargo/aws-prompt/cmd"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
