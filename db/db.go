package db

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Data Storage Schema:
// Bucket<username> -*> Key<ip> -> Value<Session>

type Storage struct {
	db *bolt.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (st *Storage) UpsetAuthEvent(authInfo AuthInfo) error {
	tx, err := st.db.Begin(true)
	if err != nil {
		return err
	}

	userBucket, err := tx.CreateBucketIfNotExists([]byte(authInfo.Username))
	if err != nil {
		return err
	}

	sessionID, err := userBucket.NextSequence()
	if err != nil {
		return err
	}

	var session Session
	rawSession := userBucket.Get([]byte(authInfo.RemoteAddr))
	if rawSession == nil {
		session = NewSession(sessionID, authInfo)
	} else {
		err = json.Unmarshal(rawSession, &session)
		if err != nil {
			return err
		}

		session.Update(authInfo)
	}

	rawSession, err = json.Marshal(session)
	if err != nil {
		return err
	}

	if err := userBucket.Put([]byte(authInfo.RemoteAddr), rawSession); err != nil {
		return err
	}

	return tx.Commit()
}

func (st *Storage) GetUserSessions(username string) ([]Session, error) {
	tx, err := st.db.Begin(false)
	if err != nil {
		return nil, err
	}

	userBucket := tx.Bucket([]byte(username))
	if userBucket != nil {
		return nil, nil
	}
	var sessions []Session
	cursor := userBucket.Cursor()

	for {
		_, rawSession := cursor.Next()
		if rawSession == nil {
			break
		}

		var session Session
		err = json.Unmarshal(rawSession, &session)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}
