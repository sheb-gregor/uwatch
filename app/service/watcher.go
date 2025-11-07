package service

import (
	"time"

	"github.com/hpcloud/tail"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sheb-gregor/uwatch/logparser"
	"github.com/sheb-gregor/uwatch/models"
	"github.com/sirupsen/logrus"
)

type Watcher struct {
	config config.Config
	hubBus EventBus
	keeper db.Storekeeper
	logger *logrus.Entry

	startTime time.Time
	logSynced bool
}

func NewWatcher(cfg config.Config, keeper db.Storekeeper, hubBus EventBus, logger *logrus.Entry) *Watcher {
	return &Watcher{
		config:    cfg,
		keeper:    keeper,
		hubBus:    hubBus,
		startTime: time.Now(),
		logger: logger.
			WithField("appLayer", "workers").
			WithField("worker", config.WWatcher)}
}

func (watcher *Watcher) Init() error {
	return nil
}

func (watcher *Watcher) Run(ctx uwe.Context) error {
	t, err := tail.TailFile(watcher.config.AuthLog, tail.Config{
		Location:  &tail.SeekInfo{Offset: 0, Whence: 0},
		MustExist: true, Follow: true, ReOpen: true})
	if err != nil {
		watcher.logger.WithError(err).Error("failed to tail auth log")
		return err
	}

	watcher.logger.Info("start event loop")
	for {
		select {
		case <-watcher.hubBus.MessageBus():
		case line := <-t.Lines:
			if line == nil {
				continue
			}
			watcher.logger.Debug("new auth log line")

			authInfo, err := logparser.ParseLine(line.Text)
			if err != nil {
				watcher.logger.WithError(err).Debug("invalid auth log line")
				continue
			}

			if authInfo.Date.Before(watcher.keeper.Stats.LastLogIn) {
				continue
			}

			session, err := watcher.keeper.StoreAuthEvent(*authInfo)
			if err != nil {
				watcher.logger.WithError(err).Error("SetAuthEvent failed")
				continue
			}

			if authInfo.Status != models.AuthAccepted {
				continue
			}

			if !watcher.logSynced {
				watcher.logSynced = authInfo.Date.After(watcher.startTime)
				continue
			}

			_ = watcher.hubBus.SendMessage(config.WTGBot, session)
			watcher.logger.Debug("broadcast session to bots")
		case <-ctx.Done():
			watcher.logger.Info("finish event loop")
			return nil
		}

	}
}
