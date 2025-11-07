package main

import (
	"flag"

	"github.com/sheb-gregor/uwatch/app"
	"github.com/sheb-gregor/uwatch/config"
)

func main() {
	var configPath = flag.String(
		"config",
		"./config.yaml",
		"path to configuration file in JSON or YAML format")

	flag.Parse()
	cfg := config.GetConfig(*configPath)

	app.Run(cfg)
}
