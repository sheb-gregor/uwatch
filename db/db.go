package db

import (
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

type StorageI interface {
	Auth() AuthStorage
	TG() TGStorage
	Slack() SlackStorage
}

// Auth Storage Schema:
// Bucket<username> -*> Key<ip> -> Value<Session>
type AuthStorage interface {
	UpsetAuthEvent(authInfo AuthInfo) (Session, error)
	GetUserSessions(username string) ([]Session, error)
}

type TGStorage interface {
	AddUser(username string, chatID int64) error
	Mute(username string, chatID int64) error
	GetUsers() (map[string]TGChatInfo, error)
	GetUser(username string) (TGChatInfo, error)
}

type SlackStorage interface {
}

type Storage struct {
	authDB *bolt.DB
	tgDB   *bolt.DB
}

func NewStorage(dbPath string) (StorageI, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err = os.Mkdir(dbPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	authDB, err := bolt.Open(dbPath+"/auth.db", 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	tgDB, err := bolt.Open(dbPath+"/tg.db", 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &Storage{authDB: authDB, tgDB: tgDB}, nil
}

func (st *Storage) Auth() AuthStorage {
	return &authStorage{
		db: st.authDB,
	}
}

func (st *Storage) TG() TGStorage {
	return &tgStorage{
		db: st.tgDB,
	}
}

func (st *Storage) Slack() SlackStorage {
	// todo:
	return nil
}
