package db

import "time"

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
	ID             uint64           `json:"session_id"`
	Username       string           `json:"username"`
	AuthMethods    map[string]int32 `json:"auth_methods,omitempty"`
	RemoteAddr     string           `json:"remote_addr"`
	ConnsCount     int32            `json:"conns_count"`
	FirstLogInTime time.Time        `json:"login_time,omitempty"`
	LastLogInTime  time.Time        `json:"login_time,omitempty"`
	LastLogOutTime *time.Time       `json:"logout_time,omitempty"`
}

func NewSession(sessionID uint64, authInfo AuthInfo) Session {
	return Session{
		ID:             sessionID,
		Username:       authInfo.Username,
		AuthMethods:    map[string]int32{authInfo.AuthMethod: 1},
		RemoteAddr:     authInfo.RemoteAddr,
		ConnsCount:     1,
		FirstLogInTime: authInfo.Date,
		LastLogInTime:  authInfo.Date,
		LastLogOutTime: nil,
	}
}

func (s *Session) Update(info AuthInfo) {
	switch info.Status {
	case AuthAccepted:
		s.ConnsCount = s.ConnsCount + 1
		s.AuthMethods[info.AuthMethod] += 1
		s.LastLogInTime = info.Date
	case AuthDisconnected:
		s.ConnsCount -= 1
		s.AuthMethods[info.AuthMethod] -= 1

		s.LastLogOutTime = &time.Time{}
		*s.LastLogOutTime = info.Date
	}

}
