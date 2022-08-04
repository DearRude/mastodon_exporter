package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/peterbourgon/ff/v3"
)

type Config struct {
	InstanceURI   string
	CheckInterval time.Duration
	StderrLogger  log.Logger
}

func GenConfig() Config {
	log.Printf("Read configurations.")
	fs := flag.NewFlagSet("mastodon_exporter", flag.ContinueOnError)
	var (
		instanceUri   = fs.String("instanceUri", "", "URI of the mastodon instance")
		checkInterval = fs.Duration("checkInterval", 30*time.Second, "Interval for check requests in go duratrion format")
		_             = fs.String("config", "", "config file (optional)")
	)
	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarNoPrefix(),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)
	if err != nil {
		log.Fatalf("Unable to parse args. Error: %s", err)
	}

	return Config{
		InstanceURI:   *instanceUri,
		CheckInterval: *checkInterval,
		StderrLogger:  *log.New(os.Stderr, "", log.LstdFlags),
	}
}
