package db

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

type AuthStatus string

const (
	AuthAccepted     AuthStatus = "Accepted"
	AuthDisconnected AuthStatus = "Disconnected"
	AuthFailed       AuthStatus = "Failed"
)

type AuthInfo struct {
	Status     AuthStatus `json:"status"`
	Username   string     `json:"username"`
	AuthMethod string     `json:"auth_method,omitempty"`
	RemoteAddr string     `json:"remote_addr"`
	Date       time.Time  `json:"date"`
}

type Session struct {
	ID          uint64           `json:"session_id"`
	Status      AuthStatus       `json:"status"`
	Username    string           `json:"username"`
	AuthMethods map[string]int32 `json:"auth_methods,omitempty"`
	RemoteAddr  string           `json:"remote_addr"`

	ConnsCount     int32      `json:"conns_count"`
	FirstLogInTime *time.Time `json:"login_time,omitempty"`
	LastLogInTime  *time.Time `json:"login_time,omitempty"`
	LastLogOutTime *time.Time `json:"logout_time,omitempty"`

	FailsCount      int32      `json:"fails_count,omitempty"`
	LastAttemptTime *time.Time `json:"last_attempt_time,omitempty"`
}

func NewSession(sessionID uint64, info AuthInfo) Session {
	s := Session{
		ID:          sessionID,
		Status:      info.Status,
		Username:    info.Username,
		AuthMethods: map[string]int32{},
		RemoteAddr:  info.RemoteAddr}

	s.Update(info)

	return s
}

func (s *Session) Update(info AuthInfo) {
	switch info.Status {
	case AuthAccepted:
		s.ConnsCount = s.ConnsCount + 1
		s.AuthMethods[info.AuthMethod] += 1

		if s.FirstLogInTime == nil {
			s.FirstLogInTime = &info.Date
		}

		s.LastLogInTime = &info.Date
	case AuthDisconnected:
		if s.ConnsCount > 0 {
			s.ConnsCount -= 1
		}
		if s.AuthMethods[info.AuthMethod] > 0 {
			s.AuthMethods[info.AuthMethod] -= 1
		}

		s.LastLogOutTime = &info.Date
	case AuthFailed:
		s.FailsCount += 1
		s.LastAttemptTime = &info.Date
	}

	s.Status = info.Status
}

type authStorage struct {
	db *bolt.DB
}

func (st *authStorage) UpsetAuthEvent(authInfo AuthInfo) (session Session, err error) {
	tx, err := st.db.Begin(true)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	userBucket, err := tx.CreateBucketIfNotExists([]byte(authInfo.Username))
	if err != nil {
		return
	}

	sessionID, err := userBucket.NextSequence()
	if err != nil {
		return
	}

	rawSession := userBucket.Get([]byte(authInfo.RemoteAddr))
	if rawSession == nil {
		session = NewSession(sessionID, authInfo)
	} else {
		err = json.Unmarshal(rawSession, &session)
		if err != nil {
			return
		}

		session.Update(authInfo)
	}

	rawSession, err = json.Marshal(session)
	if err != nil {
		return
	}

	if err = userBucket.Put([]byte(authInfo.RemoteAddr), rawSession); err != nil {
		return
	}

	return
}

func (st *authStorage) GetUserSessions(username string) (sessions []Session, err error) {
	tx, err := st.db.Begin(false)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	userBucket := tx.Bucket([]byte(username))
	if userBucket != nil {
		return
	}

	cursor := userBucket.Cursor()

	for {
		_, rawSession := cursor.Next()
		if rawSession == nil {
			break
		}

		var session Session
		err = json.Unmarshal(rawSession, &session)
		if err != nil {
			return
		}
		sessions = append(sessions, session)
	}

	return
}
