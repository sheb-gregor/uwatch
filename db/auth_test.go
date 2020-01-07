package db

import (
	"reflect"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestNewSession(t *testing.T) {
	type args struct {
		sessionID uint64
		authInfo  AuthInfo
	}
	tests := []struct {
		name string
		args args
		want Session
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSession(tt.args.sessionID, tt.args.authInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSession() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_authStorage_GetUserSessions(t *testing.T) {
	type fields struct {
		db *bolt.DB
	}
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &authStorage{
				db: tt.fields.db,
			}
			got, err := st.GetUserSessions(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserSessions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserSessions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
