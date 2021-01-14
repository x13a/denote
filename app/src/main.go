package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"bitbucket.org/x31a/denote/app/src/denote"
	"bitbucket.org/x31a/denote/app/src/denote/config"
)

const (
	ExitSuccess = 0
	ExitUsage   = 2
)

type Opts struct {
	config config.Config
}

func getOpts() *Opts {
	opts := &Opts{}
	isVersion := flag.Bool("V", false, "Print version and exit")
	flag.Parse()
	if *isVersion {
		fmt.Println(denote.Version)
		os.Exit(ExitSuccess)
	}
	if err := opts.config.FromEnv(); err != nil {
		fmt.Fprintln(flag.CommandLine.Output(), err)
		os.Exit(ExitUsage)
	}
	return opts
}

func main() {
	opts := getOpts()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer cancel()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		defer signal.Stop(sigChan)
		<-sigChan
	}()
	log.Printf("Listen on: %q\n", opts.config.Addr)
	if err := denote.Run(ctx, opts.config); err != nil {
		log.Fatalln(err)
	}
}
