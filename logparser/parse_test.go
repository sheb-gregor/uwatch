package logparser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sheb-gregor/uwatch/db"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		logLine string
		want    *db.AuthInfo
		wantErr bool
	}{
		{
			wantErr: false,
			want:    &db.AuthInfo{Status: db.AuthAccepted, Username: "sheb", AuthMethod: "publickey", RemoteAddr: "188.163.50.118"},
			logLine: "Jan  6 14:07:25 teamo sshd[31215]: Accepted publickey for sheb from 188.163.50.118 port 11087 ssh2: RSA SHA256:dKBV5Ama80sfH1e3G03VQ92kfUtQvn67zh4ebLm7smw",
		},
		{
			wantErr: false,
			want:    &db.AuthInfo{Status: db.AuthDisconnected, Username: "sheb", AuthMethod: "", RemoteAddr: "188.163.50.118"},
			logLine: "Jan  6 14:07:25 teamo sshd[31215]: Disconnected from user sheb 188.163.50.118 port 11323",
		},
		{
			wantErr: false,
			want:    &db.AuthInfo{Status: db.AuthFailed, Username: "yro", AuthMethod: "password", RemoteAddr: "213.91.179.246"},
			logLine: "Jan  6 14:07:25 teamo sshd[31215]: Failed password for invalid user yro from 213.91.179.246 port 37353 ssh2",
		},
		{
			wantErr: false,
			want:    &db.AuthInfo{Status: db.AuthFailed, Username: "root", AuthMethod: "password", RemoteAddr: "218.92.0.164"},
			logLine: "Jan  6 14:07:25 teamo sshd[31215]: Failed password for root from 218.92.0.164 port 26493 ssh2",
		},
		{
			wantErr: true,
			logLine: "Jan  6 14:08:21 teamo sudo: pam_unix(sudo:session): session closed for user root",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("#%d", i+1), func(t *testing.T) {
			got, err := ParseLine(tt.logLine)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			assertField(t, got.AuthMethod, tt.want.AuthMethod)
			assertField(t, got.RemoteAddr, tt.want.RemoteAddr)
			assertField(t, got.Username, tt.want.Username)
			assertField(t, got.Status, tt.want.Status)
		})
	}
}

func assertField(t *testing.T, got, want interface{}) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("assertField: got = %v, want %v", got, want)
	}
}
