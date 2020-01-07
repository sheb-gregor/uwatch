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

	hub := workers.NewEventHub(10)
	storage, err := db.NewStorage(cfg.DB)
	if err != nil {
		entry.WithError(err).Fatal("unable to init storage")
		return
	}

	chief := uwe.NewChief()
	chief.UseDefaultRecover()

	chief.AddWorker(workers.WLogWatcher,
		workers.NewWatcher(cfg, hub.AddWorker(workers.WLogWatcher), entry))

	chief.AddWorker(workers.WAuthSaver,
		workers.NewAuthSaver(storage, hub.AddWorker(workers.WAuthSaver), entry))

	if cfg.TG != nil {
		chief.AddWorker(workers.WTGBot,
			workers.NewTgBot(*cfg.TG, storage, hub.AddWorker(workers.WTGBot), entry))
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
