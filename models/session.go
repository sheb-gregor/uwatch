package models

import "time"

type Session struct {
	ID          uint64           `json:"session_id" yaml:"id"`
	Status      AuthStatus       `json:"status" yaml:"status"`
	Username    string           `json:"username" yaml:"username"`
	AuthMethods map[string]int32 `json:"auth_methods,omitempty" yaml:"auth_methods"`
	RemoteAddr  string           `json:"remote_addr" yaml:"remote_addr"`
	ConnsCount  int32            `json:"conns_count" yaml:"conns_count"`

	FirstLogInTime *time.Time `json:"login_time,omitempty" yaml:"first_log_in_time"`
	LastLogInTime  *time.Time `json:"login_time,omitempty" yaml:"last_log_in_time"`
	LastLogOutTime *time.Time `json:"logout_time,omitempty" yaml:"last_log_out_time"`

	FailsCount      int32      `json:"fails_count,omitempty" yaml:"fails_count"`
	LastAttemptTime *time.Time `json:"last_attempt_time,omitempty" yaml:"last_attempt_time"`
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
