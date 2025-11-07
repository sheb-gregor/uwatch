package db

import "github.com/sheb-gregor/uwatch/models"

type Storekeeper struct {
	DB    StorageI
	Stats models.AuthStats
}

func NewStorekeeper(dbPath string) (Storekeeper, error) {
	storage, err := NewStorage(dbPath)
	if err != nil {
		return Storekeeper{}, err
	}

	stats, err := storage.Auth().GetStats()
	if err != nil {
		return Storekeeper{}, err
	}

	return Storekeeper{DB: storage, Stats: stats}, nil
}

func (keeper *Storekeeper) StoreAuthEvent(info models.AuthInfo) (models.Session, error) {
	keeper.Stats.Update(info)

	session, err := keeper.DB.Auth().GetUserSession(info.Username, info.RemoteAddr)
	if err != nil {
		return session, err
	}
	session.Update(info)

	err = keeper.DB.Auth().SetStats(keeper.Stats)
	if err != nil {
		return session, err
	}

	err = keeper.DB.Auth().SetSession(session)
	if err != nil {
		return session, err
	}

	return session, nil
}
