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
	AddUser(username string, chatID int64) error
	GetUser(username string) (TGChatInfo, error)
	Mute(username string, chatID int64) error
	GetUsers() (map[string]TGChatInfo, error)
}

type tgStorage struct {
	db *bolt.DB
}

func (st *tgStorage) AddUser(username string, chatID int64) (err error) {
	tx, err := st.db.Begin(true)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	userBucket, err := tx.CreateBucketIfNotExists([]byte(bucketTGWhitelist))
	if err != nil {
		return
	}

	info := TGChatInfo{ChatID: chatID, Muted: false}
	value, err := json.Marshal(info)
	if err != nil {
		return
	}

	err = userBucket.Put([]byte(username), value)
	if err != nil {
		return
	}

	return
}

func (st *tgStorage) Mute(username string, chatID int64) (err error) {
	tx, err := st.db.Begin(true)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	userBucket, err := tx.CreateBucketIfNotExists([]byte(bucketTGWhitelist))
	if err != nil {
		return
	}

	info := TGChatInfo{ChatID: chatID, Muted: true}
	value, err := json.Marshal(info)
	if err != nil {
		return
	}

	err = userBucket.Put([]byte(username), value)
	if err != nil {
		return
	}

	return
}

func (st *tgStorage) GetUsers() (users map[string]TGChatInfo, err error) {

	tx, err := st.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	bucket, err := tx.CreateBucketIfNotExists([]byte(bucketTGWhitelist))
	if err != nil {
		return
	}
	if bucket == nil {
		return
	}

	users = map[string]TGChatInfo{}
	cursor := bucket.Cursor()
	for user, value := cursor.First(); user != nil; user, value = cursor.Next() {
		info := TGChatInfo{}
		err = json.Unmarshal(value, &info)

		users[string(user)] = info

	}
	return users, nil
}

func (st *tgStorage) GetUser(username string) (info TGChatInfo, err error) {
	tx, err := st.db.Begin(true)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	bucket := tx.Bucket([]byte(bucketTGWhitelist))
	if bucket == nil {
		return
	}

	val := bucket.Get([]byte(username))
	err = json.Unmarshal(val, &info)

	return
}
