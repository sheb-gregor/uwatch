package service

import (
	"github.com/hpcloud/tail"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sheb-gregor/uwatch/logparser"
	"github.com/sheb-gregor/uwatch/models"
	"github.com/sirupsen/logrus"
)

type Watcher struct {
	config  config.Config
	hubBus  EventBus
	storage db.StorageI
	logger  *logrus.Entry
}

func NewWatcher(cfg config.Config, storage db.StorageI, hubBus EventBus, logger *logrus.Entry) *Watcher {
	return &Watcher{
		config:  cfg,
		storage: storage,
		hubBus:  hubBus,
		logger: logger.
			WithField("appLayer", "workers").
			WithField("worker", config.WWatcher)}
}

func (w *Watcher) Init() error {
	return nil
}

func (w *Watcher) Run(ctx uwe.Context) error {
	t, err := tail.TailFile(w.config.AuthLog, tail.Config{
		Location:  &tail.SeekInfo{Offset: 0, Whence: 0},
		MustExist: true, Follow: true, ReOpen: true})
	if err != nil {
		w.logger.WithError(err).Error("failed to tail auth log")
		return err
	}

	w.logger.Info("start event loop")
	for {
		select {
		case <-w.hubBus.MessageBus():
		case line := <-t.Lines:
			if line == nil {
				continue
			}
			w.logger.Debug("new auth log line")

			authInfo, err := logparser.ParseLine(line.Text)
			if err != nil {
				w.logger.WithError(err).Debug("invalid auth log line")
				continue
			}

			if authInfo.Status == models.AuthFailed && w.config.IgnoreFails {
				continue
			}

			session, err := w.storage.Auth().SetAuthEvent(*authInfo)
			if err != nil {
				w.logger.WithError(err).Error("SetAuthEvent failed")
				continue
			}

			_ = w.hubBus.SendMessage(config.WTGBot, session)
			w.logger.Debug("broadcast session to bots")
		case <-ctx.Done():
			w.logger.Info("finish event loop")
			return nil
		}

	}
}
