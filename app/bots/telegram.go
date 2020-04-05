package bots

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/lancer-kit/uwe/v2"
	"github.com/sheb-gregor/uwatch/app/service"
	"github.com/sheb-gregor/uwatch/config"
	"github.com/sheb-gregor/uwatch/db"
	"github.com/sheb-gregor/uwatch/models"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type TgBot struct {
	logger *logrus.Entry
	config config.TGConfig
	server string

	bot     *tgbotapi.BotAPI
	hubBus  service.EventBus
	storage db.StorageI

	allowedUsers map[string]struct{}
	users        map[string]db.TGChatInfo
}

func NewTgBot(host string, cfg config.TGConfig, storage db.StorageI, hubBus service.EventBus, logger *logrus.Entry) *TgBot {
	return &TgBot{

		storage:      storage,
		hubBus:       hubBus,
		allowedUsers: cfg.AllowedUsers,
		users:        map[string]db.TGChatInfo{},

		config: cfg,
		server: host,
		logger: logger.
			WithField("app_layer", "bots").
			WithField("worker", config.WTGBot),
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
			if msg.Sender != config.WWatcher {
				continue
			}

			tg.logger.WithField("msg_data", fmt.Sprintf("%+v", msg.Data)).
				Debug("got new msg")

			session, ok := msg.Data.(models.Session)
			if !ok {
				tg.logger.WithField("msg_data_type", fmt.Sprintf("%T", msg.Data)).
					Debug("incoming msg not a db.Session")
				continue
			}

			if session.Status != models.AuthAccepted {
				continue
			}

			rawSession, err := yaml.Marshal(session)
			if err != nil {
				tg.logger.WithError(err).Error("unable to MarshalIndent session")
				continue
			}

			for user, info := range tg.users {
				if info.Muted {
					continue
				}

				text := fmt.Sprintf(
					"Hi, %s!\n\nWe got new accepted auth at [%s] host.\n\nHere details:\n\n```\n%s\n```\n\n",
					user,
					tg.server,
					string(rawSession),
				)

				msg := tgbotapi.NewMessage(info.ChatID, text)
				msg.ParseMode = "markdown"
				msg.DisableWebPagePreview = true

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
	msg.ParseMode = "markdown"

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
			msg.Text = "To add someone to whitelist pass his telegram *username*."
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
