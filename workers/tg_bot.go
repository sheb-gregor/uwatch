package workers

import (
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sirupsen/logrus"
)

type TgBot struct {
	config  config.TGConfig
	bot     *tgbotapi.BotAPI
	hubBus  EventBus
	storage db.StorageI
	logger  *logrus.Entry

	allowedUsers map[string]struct{}
	users        map[string]db.TGChatInfo
}

func NewTgBot(config config.TGConfig, storage db.StorageI, hubBus EventBus, logger *logrus.Entry) *TgBot {
	return &TgBot{
		config:       config,
		storage:      storage,
		hubBus:       hubBus,
		allowedUsers: config.AllowedUsers,
		users:        map[string]db.TGChatInfo{},
		logger: logger.
			WithField("appLayer", "workers").
			WithField("worker", WTGBot),
	}
}

func (tg *TgBot) Init() error {
	bot, err := tgbotapi.NewBotAPI(tg.config.APIToken.Get())
	if err != nil {
		tg.logger.WithError(err).Error("failed to init bot api")
		return err
	}

	bot.Debug = false
	tg.bot = bot

	tg.logger.
		WithField("username", tg.bot.Self.UserName).
		Info("authorized on account")

	users, err := tg.storage.TG().GetUsers()
	if err != nil {
		tg.logger.WithError(err).Error("failed to GetUsers")
		return err
	}

	if users != nil {
		tg.users = users
	}

	return nil
}

func (tg *TgBot) Run(ctx uwe.Context) error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	tg.logger.Info("start event loop")

	updates, err := tg.bot.GetUpdatesChan(u)
	if err != nil {
		tg.logger.WithError(err).Error("failed to GetUpdatesChan")
		return err
	}

	for {
		select {
		case msg := <-tg.hubBus.MessageBus():
			if msg.Sender != WWatcher {
				continue
			}

			tg.logger.WithField("msg_data", fmt.Sprintf("%+v", msg.Data)).
				Debug("got new msg")

			session, ok := msg.Data.(db.Session)
			if !ok {
				tg.logger.WithField("msg_data_type", fmt.Sprintf("%T", msg.Data)).
					Debug("incoming msg not a db.Session")
				continue
			}

			if session.Status != db.AuthAccepted {
				continue
			}

			rawSession, err := json.MarshalIndent(session, "", "  ")
			if err != nil {
				tg.logger.WithError(err).Error("unable to MarshalIndent session")
				continue
			}

			for user, info := range tg.users {
				if info.Muted {
					continue
				}

				text := fmt.Sprintf(
					"Hi, %s!\n\nWe got new accepted auth at server!\n\nHere details:\n\n```\n%s\n```\n\n",
					user,
					string(rawSession),
				)
				msg := tgbotapi.NewMessage(info.ChatID, text)
				if _, err := tg.bot.Send(msg); err != nil {
					tg.logger.
						WithError(err).
						WithField("user", user).
						Error("unable to send message to user")
					continue
				}

			}

		case update := <-updates:
			tg.logger.
				Debug("new tg chat update")

			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}

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
	if _, ok := tg.users[update.Message.From.UserName]; ok {
		return true
	}

	_, ok := tg.allowedUsers[update.Message.From.UserName]
	if !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Тьфу на тебя! Не буду с тобой дружить!")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.logger.
				WithError(err).
				WithField("user", update.Message.From.UserName).
				Error("unable to send message to user")
		}
		return false
	}

	err := tg.storage.TG().AddUser(update.Message.From.UserName, update.Message.Chat.ID)
	if err != nil {
		tg.logger.
			WithError(err).
			WithField("user", update.Message.From.UserName).
			Error("unable to save user into db")
		return true
	}

	tg.users[update.Message.From.UserName] = db.TGChatInfo{ChatID: update.Message.Chat.ID}
	return true
}

func (tg *TgBot) processUpdate(update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	logger := tg.logger.
		WithField("command", update.Message.Command()).
		WithField("user", update.Message.From.UserName)

	logger.Debug("process new update")

	switch update.Message.Command() {
	case "add_to_whitelist":
		tgUser := update.Message.CommandArguments()
		if tgUser != "" {
			tg.allowedUsers[tgUser] = struct{}{}
			msg.Text = fmt.Sprintf("User @%s added to whitelist.\n Wait for messages from them.", tgUser)
		} else {
			msg.Text = "To add someone to whitelist pass his telegram <b>username</b>."
		}

	case "/status":

	case "help":
		msg.Text = botHelp
	default:
		msg.Text = botHelp
	}

	if _, err := tg.bot.Send(msg); err != nil {
		logger.
			WithError(err).
			WithField("text", msg).
			Error("unable to send message to user")
	}

}

const botHelp = `
Available commands:
	/help 	print help
	
	/add_to_whitelist <tgUsername>	add telegram account to bot white list
	/mute	disable sending new auth updates
	
	/status	show list of active sessions at server
	/status <user>	show session status for the <user> at server 
	/all_sessions	show list of all sessions at server
	/all_sessions <user>	show list of all sessions for the <user> at server
`
