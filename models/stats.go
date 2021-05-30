package models

import "time"

type AuthStats struct {
	ActiveSessions Stats `json:"active_sessions" yaml:"active_sessions"`
	TotalSessions  Stats `json:"total_sessions" yaml:"total_sessions"`
	FailList       Stats `json:"fail_list" yaml:"fail_list"`

	StatsStarted time.Time `json:"stats_started" yaml:"stats_started"`
	LastLogIn    time.Time `json:"last_log_in" yaml:"last_log_in"`
}

func NewAuthStats() AuthStats {
	return AuthStats{
		ActiveSessions: NewStats(),
		TotalSessions:  NewStats(),
		StatsStarted:   time.Now(),
		LastLogIn:      time.Time{},
	}
}

func (stats *AuthStats) Update(info AuthInfo) {
	if info.Date.Before(stats.StatsStarted) {
		stats.StatsStarted = info.Date
	}

	switch info.Status {
	case AuthAccepted:
		if info.Date.After(stats.LastLogIn) {
			stats.LastLogIn = info.Date
		}

		stats.ActiveSessions.Increment(info.Username, info.RemoteAddr)
		stats.TotalSessions.Increment(info.Username, info.RemoteAddr)
	case AuthDisconnected:
		stats.ActiveSessions.Decrement(info.Username, info.RemoteAddr)
	case AuthFailed:
		stats.FailList.Increment(info.Username, info.RemoteAddr)
	}
}

type Stats struct {
	Count    int64                       `json:"count" yaml:"count"`
	ByUsers  map[string]int64            `json:"by_users" yaml:"by_users"`
	ByRemote map[string]map[string]int64 `json:"by_remote" yaml:"by_remote"`
}

func NewStats() Stats {
	return Stats{
		Count:    0,
		ByUsers:  map[string]int64{},
		ByRemote: map[string]map[string]int64{},
	}
}

func (stats *Stats) Increment(user, remote string) {
	stats.Count += 1

	byUserCount := stats.ByUsers[user]
	stats.ByUsers[user] = byUserCount + 1

	byRemote := stats.ByRemote[remote]
	if byRemote == nil {
		byRemote = map[string]int64{}
	}

	byUserCount = byRemote[user]
	byRemote[user] = byUserCount + 1
	stats.ByRemote[remote] = byRemote
}

func (stats *Stats) Decrement(user, remote string) {
	if stats.Count > 0 {
		stats.Count -= 1
	}

	byUserCount := stats.ByUsers[user]
	if byUserCount > 0 {
		byUserCount -= 1
	}
	stats.ByUsers[user] = byUserCount

	byRemote := stats.ByRemote[remote]
	if byRemote == nil {
		byRemote = map[string]int64{}
	}

	byUserCount = byRemote[user]
	if byUserCount > 0 {
		byUserCount -= 1
	}

	byRemote[user] = byUserCount
	stats.ByRemote[remote] = byRemote
}
