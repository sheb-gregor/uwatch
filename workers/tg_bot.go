package workers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
)

type TgBot struct {
	config  config.TGConfig
	bot     *tgbotapi.BotAPI
	hubBus  EventBus
	storage *db.Storage
}

func NewTgBot(config config.TGConfig, storage *db.Storage, hubBus EventBus) *TgBot {
	return &TgBot{config: config, storage: storage, hubBus: hubBus}
}

func (tg *TgBot) Init() error {
	bot, err := tgbotapi.NewBotAPI(tg.config.APIToken.Get())
	if err != nil {
		return err
	}

	bot.Debug = false
	tg.bot = bot

	return nil
}

func (tg *TgBot) Run(ctx uwe.Context) error {
	log.Printf("Authorized on account %s", tg.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := tg.bot.GetUpdatesChan(u)
	for {
		select {
		case msg := <-tg.hubBus.MessageBus():
			log.Print(msg.Data)

		case update := <-updates:
			if !tg.verifyAuth(update) {
				continue
			}

			tg.processUpdate(update)
		case <-ctx.Done():
			tg.bot.StopReceivingUpdates()
			return nil
		}
	}
}

func (tg *TgBot) verifyAuth(update tgbotapi.Update) bool {
	if _, ok := tg.config.AllowedUsers[update.Message.From.UserName]; !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Тьфу на тебя! Не буду с тобой дружить!")

		if _, err := tg.bot.Send(msg); err != nil {
			log.Print(err)
		}
		return false
	}

	return true
}

func (tg *TgBot) processUpdate(update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	// Create a new MessageConfig. We don't have text yet,
	// so we should leave it empty.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	// Extract the command from the Message.
	switch update.Message.Command() {
	case "help":
		msg.Text = "type /sayhi or /status."
	case "sayhi":
		msg.Text = "Hi :) => " + update.Message.From.UserName
	case "status":
		msg.Text = "I'm ok."
	default:
		msg.Text = "I don't know that command"
	}

	if _, err := tg.bot.Send(msg); err != nil {
		log.Panic(err)
	}

}
