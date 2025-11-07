package db

import (
	"time"

	bolt "go.etcd.io/bbolt"
)

type StorageI interface {
	Auth() AuthStorage
	TG() TGStorage
	Slack() SlackStorage
}

type SlackStorage interface {
}

type Storage struct {
	db *bolt.DB
}

func NewStorage(dbPath string) (StorageI, error) {
	db, err := bolt.Open(dbPath, 0644, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (st *Storage) Auth() AuthStorage {
	return &authStorage{
		db: st.db,
	}
}

func (st *Storage) TG() TGStorage {
	return &tgStorage{
		db: st.db,
	}
}

func (st *Storage) Slack() SlackStorage {
	// todo:
	return nil
}
