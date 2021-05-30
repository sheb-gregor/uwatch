package db

import (
	"encoding/json"

	"github.com/sheb-gregor/uwatch/models"
	bolt "go.etcd.io/bbolt"
)

const (
	bucketAuthSessions = "auth_sessions"
	bucketAuthStats    = "auth_stats"
)

// AuthStorage Schema:
// Bucket<username> -*> Key<ip> -> Value<Session>
type AuthStorage interface {
	GetStats() (models.AuthStats, error)
	SetStats(stats models.AuthStats) error
	SetSession(session models.Session) error
	GetSessions() (map[string]map[string]models.Session, error)
	GetUserSession(username, remote string) (models.Session, error)
	GetUserSessions(username string) (map[string]models.Session, error)
}

type authStorage struct {
	db *bolt.DB
}

func (st *authStorage) GetStats() (models.AuthStats, error) {
	stats := models.NewAuthStats()

	err := st.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketAuthStats))
		if bucket == nil {
			return nil
		}

		rawStats := bucket.Get([]byte(bucketAuthStats))
		if rawStats == nil {
			stats = models.NewAuthStats()
			return nil
		}

		return json.Unmarshal(rawStats, &stats)
	})

	return stats, err
}

func (st *authStorage) SetStats(stats models.AuthStats) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketAuthStats))
		if err != nil {
			return err
		}

		rawStats, err := json.Marshal(stats)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(bucketAuthStats), rawStats)
	})

}

func (st *authStorage) SetSession(session models.Session) error {
	return st.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketAuthSessions))
		if err != nil {
			return err
		}

		userBucket, err := bucket.CreateBucketIfNotExists([]byte(session.Username))
		if err != nil {
			return err
		}

		rawSession, err := json.Marshal(session)
		if err != nil {
			return err
		}

		return userBucket.Put([]byte(session.RemoteAddr), rawSession)
	})
}

func (st *authStorage) GetUserSessions(username string) (map[string]models.Session, error) {
	var sessions = map[string]models.Session{}

	err := st.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketAuthSessions))
		if bucket == nil {
			return nil
		}

		userBucket := bucket.Bucket([]byte(username))
		if userBucket == nil {
			return nil
		}

		cursor := userBucket.Cursor()
		for remote, value := cursor.First(); remote != nil; remote, value = cursor.Next() {
			var session models.Session
			err := json.Unmarshal(value, &session)
			if err != nil {
				return err
			}

			sessions[string(remote)] = session
		}

		return nil
	})

	return sessions, err
}

func (st *authStorage) GetUserSession(username, remote string) (models.Session, error) {
	var session models.Session
	session.Username = username
	session.RemoteAddr = remote

	err := st.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketAuthSessions))
		if bucket == nil {
			return nil
		}

		userBucket := bucket.Bucket([]byte(username))
		if userBucket == nil {
			return nil
		}

		value := userBucket.Get([]byte(remote))
		if value == nil {
			return nil
		}

		return json.Unmarshal(value, &session)
	})

	return session, err
}

func (st *authStorage) GetSessions() (map[string]map[string]models.Session, error) {
	sessions := map[string]map[string]models.Session{}

	stats, err := st.GetStats()
	if err != nil {
		return nil, err
	}

	for user, count := range stats.ActiveSessions.ByUsers {
		if count < 1 {
			continue
		}

		sessions[user], err = st.GetUserSessions(user)
		if err != nil {
			return nil, err
		}
	}

	return sessions, err
}
