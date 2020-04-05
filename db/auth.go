package db

import (
	"encoding/json"

	"github.com/sheb-gregor/uwatch/models"
	bolt "go.etcd.io/bbolt"
)

const (
	bucketAuthData  = "auth_data"
	bucketAuthStats = "stats"
)

// Auth Storage Schema:
// Bucket<username> -*> Key<ip> -> Value<Session>
type AuthStorage interface {
	GetStats() (models.Stats, error)
	UpdateStats(authInfo models.Stats) error
	SetAuthEvent(session models.AuthInfo) error
	SetSession(session models.Session) error
	GetSessions() (map[string]map[string]models.Session, error)
	GetUserSessions(username string) (map[string]models.Session, error)
}

type authStorage struct {
	db *bolt.DB
}

func (st *authStorage) GetStats() (models.Stats, error) {
	panic("implement me")
}

func (st *authStorage) UpdateStats(authInfo models.Stats) error {
	panic("implement me")
}

func (st *authStorage) SetAuthEvent(authInfo models.AuthInfo) (err error) {
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

	userBucket, err := tx.CreateBucketIfNotExists([]byte(authInfo.Username))
	if err != nil {
		return
	}

	sessionID, err := userBucket.NextSequence()
	if err != nil {
		return
	}

	var session models.Session

	rawSession := userBucket.Get([]byte(authInfo.RemoteAddr))
	if rawSession == nil {
		session = models.NewSession(sessionID, authInfo)
	} else {
		err = json.Unmarshal(rawSession, &session)
		if err != nil {
			return
		}

		session.Update(authInfo)
	}

	rawSession, err = json.Marshal(session)
	if err != nil {
		return
	}

	if err = userBucket.Put([]byte(authInfo.RemoteAddr), rawSession); err != nil {
		return
	}

	return
}

func (st *authStorage) SetSession(session models.Session) (err error) {
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

	// userBucket, err := tx.CreateBucketIfNotExists([]byte(authInfo.Username))
	// if err != nil {
	// 	return
	// }
	//
	// sessionID, err := userBucket.NextSequence()
	// if err != nil {
	// 	return
	// }
	//
	// rawSession := userBucket.Get([]byte(authInfo.RemoteAddr))
	// if rawSession == nil {
	// 	session = models.NewSession(sessionID, authInfo)
	// } else {
	// 	err = json.Unmarshal(rawSession, &session)
	// 	if err != nil {
	// 		return
	// 	}
	//
	// 	session.Update(authInfo)
	// }
	//
	// rawSession, err = json.Marshal(session)
	// if err != nil {
	// 	return
	// }
	//
	// if err = userBucket.Put([]byte(authInfo.RemoteAddr), rawSession); err != nil {
	// 	return
	// }

	return
}

func (st *authStorage) GetSessions() (sessions map[string]map[string]models.Session, err error) {
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

	authDataBucket := tx.Bucket([]byte(bucketAuthData))
	if authDataBucket == nil {
		return
	}

	// authDataBucket.Bucket()
	//
	// sessions = map[string]models.Session{}
	// cursor := authDataBucket.Cursor()
	// for remote, value := cursor.First(); remote != nil; remote, value = cursor.Next() {
	// 	var session models.Session
	// 	err = json.Unmarshal(value, &session)
	// 	if err != nil {
	// 		return
	// 	}
	//
	// 	sessions[string(remote)] = session
	// }

	return
}

func (st *authStorage) GetUserSessions(username string) (sessions map[string]models.Session, err error) {
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

	bucket := tx.Bucket([]byte(username))
	if bucket == nil {
		return
	}

	sessions = map[string]models.Session{}
	cursor := bucket.Cursor()
	for remote, value := cursor.First(); remote != nil; remote, value = cursor.Next() {
		var session models.Session
		err = json.Unmarshal(value, &session)
		if err != nil {
			return
		}

		sessions[string(remote)] = session
	}

	return
}
