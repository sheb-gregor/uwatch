package db

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

const bucketTGWhitelist = "tg_whitelist"

type TGChatInfo struct {
	ChatID int64
	Muted  bool
}

type TGStorage interface {
	AddUser(username string, chatID int64, mute bool) error
	GetUser(username string) (TGChatInfo, error)
	GetUsers() (map[string]TGChatInfo, error)
}

type tgStorage struct {
	db *bolt.DB
}

func (st *tgStorage) AddUser(username string, chatID int64, mute bool) (err error) {
	return st.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketTGWhitelist))
		if bucket == nil {
			return nil
		}

		info := TGChatInfo{ChatID: chatID, Muted: mute}
		value, err := json.Marshal(info)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(username), value)
	})
}

func (st *tgStorage) GetUser(username string) (TGChatInfo, error) {
	var info TGChatInfo
	err := st.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTGWhitelist))
		if bucket == nil {
			return nil
		}

		val := bucket.Get([]byte(username))
		return json.Unmarshal(val, &info)
	})

	return info, err
}

func (st *tgStorage) GetUsers() (map[string]TGChatInfo, error) {
	var users = map[string]TGChatInfo{}
	err := st.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketTGWhitelist))

		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()
		for user, value := cursor.First(); user != nil; user, value = cursor.Next() {
			info := TGChatInfo{}
			err := json.Unmarshal(value, &info)
			if err != nil {
				return err
			}

			users[string(user)] = info
		}
		return nil
	})

	return users, err
}
