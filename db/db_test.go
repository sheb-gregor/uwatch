package db

import (
	"os"
	"testing"

	"github.com/sheb-gregor/uwatch/config"
	. "github.com/smartystreets/goconvey/convey"
)

// func TestStorage_Auth(t *testing.T) {
// 	Convey("All methods of Auth Storage works as expected", t, func() {
// 		const tempPath = "./tmp_1"
// 		cfg := config.Config{DataDir: tempPath}
// 		err := cfg.Init()
// 		So(err, ShouldBeNil)
//
// 		storage, err := NewStorage(cfg.GetPath("db"))
// 		So(err, ShouldBeNil)
// 		So(storage, ShouldNotBeNil)
//
// 		auth := storage.Auth()
//
// 		Convey("Auth Storage allows to", func() {
// 			const usernameBob = "bob"
//
// 			authInfo := models.AuthInfo{
// 				Status:     models.AuthAccepted,
// 				Username:   usernameBob,
// 				AuthMethod: "password",
// 				RemoteAddr: "188.163.50.118",
// 				Date:       time.Now(),
// 			}
//
// 			Convey("save new session ", func() {
// 				session, err := auth.SetAuthEvent(authInfo)
// 				So(err, ShouldBeNil)
// 				So(session.Status, ShouldEqual, authInfo.Status)
// 				So(session.Username, ShouldEqual, authInfo.Username)
// 				So(session.RemoteAddr, ShouldEqual, authInfo.RemoteAddr)
//
// 				sessions, err := auth.GetUserSessions(usernameBob)
// 				So(err, ShouldBeNil)
// 				So(len(sessions), ShouldBeGreaterThan, 0)
//
// 				session, ok := sessions[authInfo.RemoteAddr]
// 				So(ok, ShouldBeTrue)
// 				So(session.Status, ShouldEqual, authInfo.Status)
// 				So(session.Username, ShouldEqual, authInfo.Username)
// 				So(session.RemoteAddr, ShouldEqual, authInfo.RemoteAddr)
// 				So(len(session.AuthMethods), ShouldEqual, 1)
// 			})
//
// 		})
//
// 		err = os.RemoveAll(tempPath)
// 		So(err, ShouldBeNil)
// 	})
// }

func TestStorage_TG(t *testing.T) {
	Convey("All methods of TG Storage works as expected", t, func() {
		const tempPath = "./tmp_2"
		cfg := config.Config{DataDir: tempPath}
		err := cfg.Init()
		So(err, ShouldBeNil)

		storage, err := NewStorage(cfg.GetPath("db"))
		So(err, ShouldBeNil)
		So(storage, ShouldNotBeNil)

		tg := storage.TG()

		Convey("TG Storage allows to", func() {
			const usernameBob = "bob"
			const chatIDBob = 42

			const usernameDave = "dave"
			const chatIDDave = 13

			err = tg.AddUser(usernameBob, chatIDBob, false)
			So(err, ShouldBeNil)

			err = tg.AddUser(usernameDave, chatIDDave, false)
			So(err, ShouldBeNil)

			Convey("add multiple users", func() {
				chatInfo, err := tg.GetUser(usernameDave)
				So(err, ShouldBeNil)
				So(chatInfo.ChatID, ShouldEqual, chatIDDave)
				So(chatInfo.Muted, ShouldBeFalse)

				users, err := tg.GetUsers()
				So(err, ShouldBeNil)
				So(len(users), ShouldEqual, 2)

				chatInfo, ok := users[usernameBob]
				So(ok, ShouldBeTrue)
				So(chatInfo.ChatID, ShouldEqual, chatIDBob)
				So(chatInfo.Muted, ShouldBeFalse)

				chatInfo, ok = users[usernameDave]
				So(ok, ShouldBeTrue)
				So(chatInfo.ChatID, ShouldEqual, chatIDDave)
				So(chatInfo.Muted, ShouldBeFalse)
			})

			Convey("mute some chat", func() {
				err = tg.AddUser(usernameDave, chatIDDave, true)
				So(err, ShouldBeNil)

				chatInfo, err := tg.GetUser(usernameDave)
				So(err, ShouldBeNil)
				So(chatInfo.ChatID, ShouldEqual, chatIDDave)
				So(chatInfo.Muted, ShouldBeTrue)

				users, err := tg.GetUsers()
				So(err, ShouldBeNil)
				So(len(users), ShouldEqual, 2)

				chatInfo, ok := users[usernameBob]
				So(ok, ShouldBeTrue)
				So(chatInfo.ChatID, ShouldEqual, chatIDBob)
				So(chatInfo.Muted, ShouldBeFalse)

				chatInfo, ok = users[usernameDave]
				So(ok, ShouldBeTrue)
				So(chatInfo.ChatID, ShouldEqual, chatIDDave)
				So(chatInfo.Muted, ShouldBeTrue)
			})

		})

		err = os.RemoveAll(tempPath)
		So(err, ShouldBeNil)
	})
}
