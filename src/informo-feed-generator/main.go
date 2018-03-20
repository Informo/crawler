package main

import (
	"flag"
	"fmt"

	"common"
	"common/config"
	"common/database"
	"informo-feed-generator/generator"

	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config", "config.yaml", "Configuration file")
	debug      = flag.Bool("debug", false, "Print debugging messages")
)

func main() {
	// Parse the command line arguments.
	flag.Parse()

	// Configure the logger.
	common.LogConfig(*debug)

	// Load the configuration from the provided configuration file.
	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(fmt.Errorf("Couldn't load config: %s", err.Error()))
	}

	// Check if there's a "feed" section in the configuration file.
	if cfg.FeedsConfig == nil {
		logrus.Panic(fmt.Errorf("No 'feeds' configuration found, please provide one"))
	}

	// Open the database and prepare the required statements.
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logrus.Panic(fmt.Errorf("Couldn't open database: %s", err.Error()))
	}

	g := generator.NewGenerator(db, cfg.FeedsConfig)

	if err = g.SetupAndServe(); err != nil {
		logrus.Panic(fmt.Errorf("Something went wrong with the web server: %s", err.Error()))
	}
}
