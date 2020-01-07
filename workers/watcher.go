package workers

import (
	"fmt"
	"log"

	"github.com/hpcloud/tail"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/logparser"
)

type Watcher struct {
	config config.Config
	hubBus EventBus
}

func NewWatcher(config config.Config, hubBus EventBus) *Watcher {
	return &Watcher{config: config, hubBus: hubBus}
}

func (w *Watcher) Init() error {
	return nil
}

func (w *Watcher) Run(ctx uwe.Context) error {
	t, err := tail.TailFile(w.config.Logfile, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		return err
	}

	for {
		select {
		case line := <-t.Lines:
			if line == nil {
				continue
			}
			fmt.Println(line.Text)

			authInfo, err := logparser.ParseLine(line.Text)
			if err != nil {
				log.Print(err)
				continue
			}

			_ = w.hubBus.SendMessage("*", *authInfo)
		case <-ctx.Done():
			return nil
		}

	}
}
