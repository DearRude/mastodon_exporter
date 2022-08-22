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
	Port          int
	CheckInterval time.Duration
	StderrLogger  log.Logger
}

func GenConfig() Config {
	log.Printf("Read configurations.")
	fs := flag.NewFlagSet("mastodon_exporter", flag.ContinueOnError)
	var (
		instanceUri   = fs.String("instanceUri", "", "URI of the mastodon instance")
		port          = fs.Int("port", 2112, "exposed port of exporter")
		checkInterval = fs.Duration("checkInterval", 30*time.Second, "Interval for check requests in go duratrion format")
	)

	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fs.String("config", "", "config file")
	} else {
		fs.String("config", ".env", "config file")
	}

	err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVars(),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.EnvParser),
	)
	if err != nil {
		log.Fatalf("Unable to parse args. Error: %s", err)
	}

	return Config{
		InstanceURI:   *instanceUri,
		Port:          *port,
		CheckInterval: *checkInterval,
		StderrLogger:  *log.New(os.Stderr, "", log.LstdFlags),
	}
}
