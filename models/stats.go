package models

import "time"

type Stats struct {
	Count    int64                       `json:"count" yaml:"count"`
	ByUsers  map[string]int64            `json:"by_users" yaml:"by_users"`
	ByRemote map[string]map[string]int64 `json:"by_remote" yaml:"by_remote"`
}

type AuthStats struct {
	ActiveSessions Stats `json:"active_sessions" yaml:"active_sessions"`
	TotalSessions  Stats `json:"total_sessions" yaml:"total_sessions"`

	StatsStarted time.Time `json:"stats_started" yaml:"stats_started"`
	LastLogIn    time.Time `json:"last_log_in" yaml:"last_log_in"`
}
