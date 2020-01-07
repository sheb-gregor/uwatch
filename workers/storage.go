package workers

import (
	"log"

	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
)

type AuthSaver struct {
	config  config.Config
	hubBus  EventBus
	storage *db.Storage
}

func NewAuthSaver(storage *db.Storage, hubBus EventBus) *AuthSaver {
	return &AuthSaver{storage: storage, hubBus: hubBus}
}

func (w *AuthSaver) Init() error {
	return nil
}

func (w *AuthSaver) Run(ctx uwe.Context) error {
	for {
		select {
		case msg := <-w.hubBus.MessageBus():
			authInfo, ok := msg.Data.(db.AuthInfo)
			if !ok {
				log.Print("not an auth info")
				continue
			}
			err := w.storage.UpsetAuthEvent(authInfo)
			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return nil
		}

	}
}
