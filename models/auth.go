package models

import "time"

type AuthStatus string

const (
	AuthAccepted     AuthStatus = "Accepted"
	AuthDisconnected AuthStatus = "Disconnected"
	AuthFailed       AuthStatus = "Failed"
)

type AuthInfo struct {
	Status     AuthStatus `json:"status" yaml:"status"`
	Username   string     `json:"username" yaml:"username"`
	AuthMethod string     `json:"auth_method,omitempty" yaml:"auth_method"`
	RemoteAddr string     `json:"remote_addr" yaml:"remote_addr"`
	Date       time.Time  `json:"date" yaml:"date"`
}
