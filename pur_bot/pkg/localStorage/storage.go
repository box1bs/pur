package localstorage

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedisStorage(dbNum int, addr, password string) (*RedisStorage, error) {
	db := &RedisStorage{
		rdb: redis.NewClient(&redis.Options{
			Addr: addr,
			Password: password,
			DB: dbNum,
		}),
		ctx: context.Background(),
	}

	if err := db.rdb.Ping(db.ctx).Err(); err != nil {
		return nil, err
	}

	return db, nil
}

func (rs *RedisStorage) SyncId(id int64) error {
	uniqId := strconv.FormatInt(id, 10)

	if err := rs.rdb.Get(rs.ctx, uniqId).Err(); err == redis.Nil {
		uid := uuid.New().String()
		log.Printf("Saving UUID: %s for key: %s", uid, uniqId)
		if err := rs.rdb.Set(rs.ctx, uniqId, uid, 0).Err(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (rs *RedisStorage) GetSyncId(id int64) (uuid.UUID, error) {
	uniqId := strconv.FormatInt(id, 10)

	uid, err := rs.rdb.Get(rs.ctx, uniqId).Result()
	if err == redis.Nil {
        return uuid.Nil, fmt.Errorf("key %s not found in Redis", uniqId)
    }
    if err != nil {
        return uuid.Nil, err
    }

	syncId, err := uuid.Parse(uid)
	if err != nil {
		return uuid.Nil, err
	}

	return syncId, nil
}