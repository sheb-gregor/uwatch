package app

import (
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/app/bots"
	"github.com/sheb-gregor/uwatch/app/service"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sirupsen/logrus"
)

func Run(cfg config.Config) {
	entry := cfg.Logger()

	storage, err := db.NewStorage(cfg.GetPath("db"))
	if err != nil {
		entry.WithError(err).Fatal("unable to init storage")
		return
	}

	chief := uwe.NewChief()
	chief.UseDefaultRecover()

	hub := service.NewEventHub(10)

	watcherBus := hub.AddWorker(config.WWatcher)
	chief.AddWorker(config.WWatcher,
		service.NewWatcher(cfg, storage, watcherBus, entry))

	if cfg.TG != nil {
		botBus := hub.AddWorker(config.WTGBot)
		chief.AddWorker(config.WTGBot,
			bots.NewTgBot(cfg.Server, *cfg.TG, storage, botBus, entry))
	}

	chief.AddWorker(config.WHub, hub)

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
