package main

import (
	"flag"
	"log"

	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sheb-gregor/uwatch/workers"
)

var configPath = flag.String("config", "./config.json", "path to configuration file")

func main() {
	flag.Parse()
	cfg := config.GetConfig(*configPath)
	hub := workers.NewEventHub(10)
	storage, err := db.NewStorage(cfg.DB)
	if err != nil {

	}

	chief := uwe.NewChief()
	chief.UseDefaultRecover()
	chief.AddWorker("log_watcher", workers.NewWatcher(cfg, hub.AddWorker("log_watcher")))
	chief.AddWorker("auth_saver", workers.NewAuthSaver(storage, hub.AddWorker("storage")))

	if cfg.TG != nil {
		chief.AddWorker("tg_bot", workers.NewTgBot(*cfg.TG, storage, hub.AddWorker("tg_bot")))
	}

	chief.AddWorker("hub", hub)

	chief.SetEventHandler(func(event uwe.Event) {
		log.Print(event)
	})

	chief.Run()
}
