// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"fmt"
	"time"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
)

const (
	appsKey = "collector:apps"
	appKey  = "collector:app:%s"
)

var (
	// ConnectRetries says how many times the client should retry a failed connection
	ConnectRetries = 5
	// ConnectRetryDelay says how long the client should wait between retries
	ConnectRetryDelay = time.Second
)

// AppStorage provides storage for applications
type AppStorage interface {
	Add(eui types.AppEUI) error
	Remove(eui types.AppEUI) error
	SetAccessKey(eui types.AppEUI, key string) error
	GetAccessKey(eui types.AppEUI) (string, error)
	List() ([]types.AppEUI, error)
	Reset() error
	Close() error
}

type redisAppStorage struct {
	client *redis.Client
}

// ConnectRedis connects to Redis using the specified options
func ConnectRedis(addr string, db int64) (AppStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	var err error
	for retries := 0; retries < ConnectRetries; retries++ {
		_, err = client.Ping().Result()
		if err == nil {
			break
		}
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		client.Close()
		return nil, err
	}
	return &redisAppStorage{client}, nil
}

func makeKey(eui types.AppEUI) string {
	return fmt.Sprintf(appKey, eui.String())
}

func (s *redisAppStorage) Add(eui types.AppEUI) error {
	return s.client.SAdd(appsKey, eui.String()).Err()
}

func (s *redisAppStorage) Remove(eui types.AppEUI) error {
	err := s.client.SRem(appsKey, eui.String()).Err()
	if err != nil {
		return err
	}
	s.client.Del(makeKey(eui))
	return nil
}

func (s *redisAppStorage) SetAccessKey(eui types.AppEUI, key string) error {
	return s.client.HSet(makeKey(eui), "key", key).Err()
}

func (s *redisAppStorage) GetAccessKey(eui types.AppEUI) (string, error) {
	m, err := s.client.HGetAllMap(makeKey(eui)).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return m["key"], nil
}

func (s *redisAppStorage) List() ([]types.AppEUI, error) {
	members, err := s.client.SMembers(appsKey).Result()
	if err != nil {
		return nil, err
	}
	euis := make([]types.AppEUI, len(members))
	for i, k := range members {
		eui, err := types.ParseAppEUI(k)
		if err != nil {
			return nil, err
		}
		euis[i] = eui
	}
	return euis, nil
}

func (s *redisAppStorage) Reset() error {
	return s.client.FlushDb().Err()
}

func (s *redisAppStorage) Close() error {
	return s.client.Close()
}
