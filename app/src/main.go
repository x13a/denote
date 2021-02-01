package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/x13a/denote/config"
	"github.com/x13a/denote/denote"
)

const (
	ExitSuccess = 0
	ExitUsage   = 2
)

func getOpts() {
	isVersion := flag.Bool("V", false, "Print version and exit")
	flag.Parse()
	if *isVersion {
		fmt.Println(denote.Version)
		os.Exit(ExitSuccess)
	}
	if err := config.LoadEnv(); err != nil {
		fmt.Fprintln(flag.CommandLine.Output(), err)
		os.Exit(ExitUsage)
	}
}

func main() {
	getOpts()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
	}()
	log.Printf("Listen on: %q\n", config.Addr)
	if err := denote.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
