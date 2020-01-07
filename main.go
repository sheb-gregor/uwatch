package main

import (
	"flag"

	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sheb-gregor/uwatch/workers"
	"github.com/sirupsen/logrus"
)

var configPath = flag.String("config", "./config.json", "path to configuration file")

func main() {
	flag.Parse()
	cfg := config.GetConfig(*configPath)

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.SetLevel(logLevel)
	entry := logger.WithField("app", "uwatch")

	storage, err := db.NewStorage(cfg.DB)
	if err != nil {
		entry.WithError(err).Fatal("unable to init storage")
		return
	}

	chief := uwe.NewChief()
	chief.UseDefaultRecover()

	hub := workers.NewEventHub(10)

	watcherBus := hub.AddWorker(workers.WWatcher)
	chief.AddWorker(workers.WWatcher,
		workers.NewWatcher(cfg, storage, watcherBus, entry))

	if cfg.TG != nil {
		botBus := hub.AddWorker(workers.WTGBot)
		chief.AddWorker(workers.WTGBot,
			workers.NewTgBot(*cfg.TG, storage, botBus, entry))
	}

	chief.AddWorker(workers.WHub, hub)

	chief.SetEventHandler(func(event uwe.Event) {
		var level logrus.Level
		switch event.Level {
		case uwe.LvlFatal, uwe.LvlError:
			level = logrus.ErrorLevel
		case uwe.LvlInfo:
			level = logrus.InfoLevel
		default:
			level = logrus.WarnLevel
		}

		entry.WithFields(event.Fields).
			Log(level, event.Message)

	})

	chief.Run()
}
