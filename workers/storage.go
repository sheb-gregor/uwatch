package workers

import (
	"fmt"

	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sirupsen/logrus"
)

type AuthSaver struct {
	config  config.Config
	hubBus  EventBus
	storage db.StorageI
	logger  *logrus.Entry
}

func NewAuthSaver(storage db.StorageI, hubBus EventBus, logger *logrus.Entry) *AuthSaver {
	return &AuthSaver{
		storage: storage,
		hubBus:  hubBus,
		logger: logger.
			WithField("appLayer", "workers").
			WithField("worker", WAuthSaver)}
}

func (w *AuthSaver) Init() error {
	return nil
}

func (w *AuthSaver) Run(ctx uwe.Context) error {
	w.logger.Info("start event loop")
	for {
		select {
		case msg := <-w.hubBus.MessageBus():
			if msg.Sender != WLogWatcher {
				continue
			}
			w.logger.WithField("msg_data", fmt.Sprintf("%+v", msg.Data)).
				Debug("got new msg")

			authInfo, ok := msg.Data.(db.AuthInfo)
			if !ok {
				w.logger.WithField("msg_data_type", fmt.Sprintf("%T", msg.Data)).
					Debug("incoming msg not a db.AuthInfo")
				continue
			}

			session, err := w.storage.Auth().UpsetAuthEvent(authInfo)
			if err != nil {
				w.logger.WithError(err).Error("UpsetAuthEvent failed")
				continue
			}

			_ = w.hubBus.SendMessage("*", session)
			w.logger.Debug("broadcast session to auth_saver")
		case <-ctx.Done():
			w.logger.Info("finish event loop")
			return nil
		}

	}
}
