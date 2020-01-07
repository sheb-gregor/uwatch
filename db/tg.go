package db

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

type TGChatInfo struct {
	ChatID int64
	Muted  bool
}

const bucketTGWhitelist = "tg_whitelist"

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
	users = map[string]TGChatInfo{}
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
	if bucket != nil {
		return
	}

	cursor := bucket.Cursor()
	for {
		user, value := cursor.Next()
		if value == nil {
			break
		}
		info := TGChatInfo{}
		err = json.Unmarshal(value, &info)

		users[string(user)] = info
	}

	return
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
	if bucket != nil {
		return
	}

	val := bucket.Get([]byte(username))
	err = json.Unmarshal(val, &info)

	return
}
