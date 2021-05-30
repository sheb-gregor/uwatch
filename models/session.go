package models

import "time"

type SessionStat struct {
	Count         int32            `json:"count" yaml:"count"`
	ByAuthMethods map[string]int32 `json:"by_auth_methods" yaml:"by_auth_methods"`
}

func (stat *SessionStat) Increment(authMethod string) {
	stat.Count += 1

	if authMethod == "" {
		return
	}

	if stat.ByAuthMethods == nil {
		stat.ByAuthMethods = map[string]int32{}
	}

	count := stat.ByAuthMethods[authMethod]
	stat.ByAuthMethods[authMethod] = count + 1
}

func (stat *SessionStat) Decrement(authMethod string) {
	if stat.Count > 0 {
		stat.Count -= 1
	}

	if authMethod == "" {
		return
	}

	if stat.ByAuthMethods == nil {
		stat.ByAuthMethods = map[string]int32{}
	}

	count := stat.ByAuthMethods[authMethod]
	if count > 0 {
		stat.ByAuthMethods[authMethod] = count - 1
	}

}

type Session struct {
	Status     AuthStatus `json:"status" yaml:"status"`
	Username   string     `json:"username" yaml:"username"`
	RemoteAddr string     `json:"remote_addr" yaml:"remote_addr"`

	Active SessionStat `json:"active"`
	Fails  SessionStat `json:"fails"`
	Total  SessionStat `json:"total"`

	FirstLogIn *time.Time `json:"first_login" yaml:"first_login"`
	LastLogIn  *time.Time `json:"last_login" yaml:"last_login"`
	LastLogOut *time.Time `json:"logout" yaml:"last_log_out"`
	LastFail   *time.Time `json:"last_fail" yaml:"last_fail"`
}

func NewSession(info AuthInfo) Session {
	s := Session{
		Status:     info.Status,
		Username:   info.Username,
		RemoteAddr: info.RemoteAddr}

	s.Update(info)

	return s
}

func (s *Session) Update(info AuthInfo) {
	switch info.Status {
	case AuthAccepted:
		s.Active.Increment(info.AuthMethod)
		s.Total.Increment(info.AuthMethod)

		if s.FirstLogIn == nil {
			s.FirstLogIn = &info.Date
		}

		s.Status = info.Status
		s.LastLogIn = &info.Date

	case AuthDisconnected:
		s.Active.Decrement(info.AuthMethod)

		if s.Active.Count == 0 {
			s.LastLogOut = &info.Date
			s.Status = info.Status
		}

	case AuthFailed:
		s.Fails.Increment(info.AuthMethod)
		s.LastFail = &info.Date

		if s.Active.Count == 0 {
			s.Status = info.Status
		}
	}

}
