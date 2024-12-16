package localstorage

import "github.com/google/uuid"

type LocalStorage interface {
	SyncId(int64) error
	GetSyncId(int64) (uuid.UUID, error)
}