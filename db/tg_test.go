package db

import (
	"reflect"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func Test_tgStorage_AddUser(t *testing.T) {
	type fields struct {
		db *bolt.DB
	}
	type args struct {
		username string
		chatID   int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &tgStorage{
				db: tt.fields.db,
			}
			if err := st.AddUser(tt.args.username, tt.args.chatID); (err != nil) != tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_tgStorage_GetUser(t *testing.T) {
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
		want    TGChatInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &tgStorage{
				db: tt.fields.db,
			}
			got, err := st.GetUser(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tgStorage_GetUsers(t *testing.T) {
	type fields struct {
		db *bolt.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]TGChatInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &tgStorage{
				db: tt.fields.db,
			}
			got, err := st.GetUsers()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUsers() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tgStorage_Mute(t *testing.T) {
	type fields struct {
		db *bolt.DB
	}
	type args struct {
		username string
		chatID   int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &tgStorage{
				db: tt.fields.db,
			}
			if err := st.Mute(tt.args.username, tt.args.chatID); (err != nil) != tt.wantErr {
				t.Errorf("Mute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
