package models

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSession(t *testing.T) {
	Convey("All methods of Session works as expected", t, func() {
		const usernameBob = "bob"
		const sessionBob = 42
		const passwordMethod = "password"

		authInfo := AuthInfo{
			Status:     AuthAccepted,
			Username:   usernameBob,
			AuthMethod: passwordMethod,
			RemoteAddr: "188.163.50.118",
			Date:       time.Now(),
		}

		session := NewSession(sessionBob, authInfo)

		So(session.Status, ShouldEqual, authInfo.Status)
		So(session.Username, ShouldEqual, authInfo.Username)
		So(session.RemoteAddr, ShouldEqual, authInfo.RemoteAddr)
		So(len(session.AuthMethods), ShouldEqual, 1)

		Convey("Auth Storage allows to", func() {

		})

	})
}
