package bots

import (
	"fmt"
	"time"

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

	bot    *tgbotapi.BotAPI
	hubBus service.EventBus
	keeper db.Storekeeper

	allowedUsers map[string]struct{}
	users        map[string]db.TGChatInfo
}

func NewTgBot(host string, cfg config.TGConfig, keeper db.Storekeeper, hubBus service.EventBus, logger *logrus.Entry) *TgBot {
	return &TgBot{
		keeper:       keeper,
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

	tg.logger.WithField("username", tg.bot.Self.UserName).
		Info("authorized on account")

	users, err := tg.keeper.DB.TG().GetUsers()
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
					Debug("incoming msg is not a db.Session")
				continue
			}

			if session.Status != models.AuthAccepted {
				continue
			}

			tg.handleNewSessionInfo(session)

		case update := <-updates:
			tg.logger.Debug("new tg chat update")

			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}

			if !tg.checkIsAuthorized(update) {
				continue
			}

			tg.processUpdate(update)
		case <-ctx.Done():
			tg.bot.StopReceivingUpdates()
			time.Sleep(time.Second)
			return nil
		}
	}
}

func (tg *TgBot) handleNewSessionInfo(session models.Session) {
	rawSession, err := yaml.Marshal(session)
	if err != nil {
		tg.logger.WithError(err).Error("unable to MarshalIndent session")
		return
	}

	for user, info := range tg.users {
		if info.Muted {
			continue
		}

		text := fmt.Sprintf(
			"Hi, %s!\n\nNew accepted auth at *%s* host.\n\nHere is details:\n```\n%s\n```\n\n",
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
}

func (tg *TgBot) checkIsAuthorized(update tgbotapi.Update) bool {
	if _, ok := tg.users[update.Message.From.UserName]; ok {
		return true
	}

	_, ok := tg.allowedUsers[update.Message.From.UserName]
	if !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ParseMode = "markdown"
		msg.Text = fmt.Sprintf("```%s```", goAway)
		if _, err := tg.bot.Send(msg); err != nil {
			tg.logger.WithError(err).
				WithField("user", update.Message.From.UserName).
				Error("unable to send message to user")
		}
		return false
	}

	err := tg.keeper.DB.TG().AddUser(update.Message.From.UserName, update.Message.Chat.ID, false)
	if err != nil {
		tg.logger.WithError(err).
			WithField("user", update.Message.From.UserName).
			Error("unable to save user into db")
		return true
	}

	tg.users[update.Message.From.UserName] = db.TGChatInfo{ChatID: update.Message.Chat.ID}
	return true
}

func (tg *TgBot) processUpdate(update tgbotapi.Update) {
	userName := update.Message.From.UserName

	logger := tg.logger.WithField("user", userName).
		WithField("command", update.Message.Command())
	logger.Debug("process new update")

	var msgText string
	switch update.Message.Command() {
	case "whitelist":
		tgUser := update.Message.CommandArguments()
		if tgUser != "" {
			tg.allowedUsers[tgUser] = struct{}{}
			msgText = fmt.Sprintf("User @%s has been whitelisted.\nWaiting for messages from him...", tgUser)
		} else {
			msgText = "To add someone to the white list, send their telegram *username*."
		}

	case "mute":
		msgText = tg.handleMute(userName, true, logger)
	case "unmute":
		msgText = tg.handleMute(userName, false, logger)
	case "stats":
		msgText = tg.handleStats(logger)
	case "status":
		msgText = tg.handleStatus(update.Message.CommandArguments(), logger)
	case "help":
		msgText = botHelp
	default:
		msgText = botHelp
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ParseMode = "markdown"
	msg.Text = msgText
	if _, err := tg.bot.Send(msg); err != nil {
		logger.WithError(err).
			WithField("text", msg).
			Error("unable to send message to user")
	}

}

func (tg *TgBot) handleMute(userName string, mute bool, logger *logrus.Entry) string {
	info := tg.users[userName]
	info.Muted = mute
	tg.users[userName] = info

	err := tg.keeper.DB.TG().AddUser(userName, info.ChatID, mute)
	if err != nil {
		logger.WithError(err).Error("unable to mute user")
		return "Action failed :("
	}

	return "Done"
}

func (tg *TgBot) handleStats(logger *logrus.Entry) string {
	stats, err := tg.keeper.DB.Auth().GetStats()
	if err != nil {
		logger.WithError(err).Error("unable to get stats")
		return "Action failed :("
	}

	rawStats, err := yaml.Marshal(stats)
	if err != nil {
		logger.WithError(err).Error("unable to marshal stats")
		return "Action failed :("
	}

	return fmt.Sprintf(
		"Actual auth statistic at *%s* host:\n```\n%s\n```\n\n",
		tg.server,
		string(rawStats))
}

func (tg *TgBot) handleStatus(user string, logger *logrus.Entry) string {
	if user == "" {
		return "Provide linux *username* to see his authorization details"
	}

	sessions, err := tg.keeper.DB.Auth().GetUserSessions(user)
	if err != nil {
		logger.WithError(err).Error("unable to get stats")
		return "Action failed :("
	}

	rawStats, err := yaml.Marshal(sessions)
	if err != nil {
		logger.WithError(err).Error("unable to marshal sessions")
		return "Action failed :("
	}

	return fmt.Sprintf(
		"Sessions info at *%s* host for *%s*:\n```\n%s\n```\n\n",
		tg.server,
		user,
		string(rawStats))

}

const botHelp = `
Available commands:
	/help 	print help
	
	/whitelist <tgUsername>	— add telegram account to bot whitelist
	/mute	— disable sending new auth updates
	/unmute	— enable sending new auth updates
	
	/stats	— show list of active sessions at server
	/status <user>	— show session info for the <user> at server
`

// /all_sessions	— show list of all sessions at server
// /all_sessions <user>	show list of all sessions for the <user> at server
// `

const goAway = `
☣☣☣ GO AWAY ☣☣☣

    \_\
   (_**)
  __) #_
 ( )...()
 || | |I|
 || | |()__/
 /\(___)
_-"""""""-_""-_
-,,,,,,,,- ,,-
|////////////////////
`
