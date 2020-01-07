package workers

import (
	"github.com/hpcloud/tail"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/logparser"
	"github.com/sirupsen/logrus"
)

type Watcher struct {
	config config.Config
	hubBus EventBus
	logger *logrus.Entry
}

func NewWatcher(config config.Config, hubBus EventBus, logger *logrus.Entry) *Watcher {
	return &Watcher{
		config: config,
		hubBus: hubBus,
		logger: logger.
			WithField("appLayer", "workers").
			WithField("worker", WLogWatcher)}
}

func (w *Watcher) Init() error {
	return nil
}

func (w *Watcher) Run(ctx uwe.Context) error {
	t, err := tail.TailFile(w.config.AuthLog, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		w.logger.WithError(err).Error("failed to tail auth log")
		return err
	}

	w.logger.Info("start event loop")
	for {
		select {
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

			_ = w.hubBus.SendMessage(WAuthSaver, *authInfo)
			w.logger.Debug("send authInfo to auth_saver")
		case <-ctx.Done():
			w.logger.Info("finish event loop")
			return nil
		}

	}
}
