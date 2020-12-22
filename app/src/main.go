package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"bitbucket.org/x31a/denote/app/src/denote"
)

const (
	ExitSuccess = 0
	ExitUsage   = 2
)

type Opts struct {
	config denote.Config
}

func getOpts() *Opts {
	opts := &Opts{}
	isHelp := flag.Bool("h", false, "Print help and exit")
	isVersion := flag.Bool("V", false, "Print version and exit")
	flag.Parse()
	if *isHelp {
		flag.Usage()
		os.Exit(ExitSuccess)
	}
	if *isVersion {
		fmt.Println(denote.Version)
		os.Exit(ExitSuccess)
	}
	if err := opts.config.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
