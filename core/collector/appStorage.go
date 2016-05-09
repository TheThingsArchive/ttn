// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"fmt"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
)

const (
	appsKey = "collector:apps"
	appKey  = "collector:app:%s"
)

// App represents a stored application
type App struct {
	EUI types.AppEUI
	Key string
}

// AppStorage provides storage for applications
type AppStorage interface {
	Add(eui types.AppEUI) error
	Remove(eui types.AppEUI) error
	SetKey(eui types.AppEUI, key string) error
	Get(eui types.AppEUI) (*App, error)
	GetAll() ([]*App, error)
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
	_, err := client.Ping().Result()
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

func (s *redisAppStorage) SetKey(eui types.AppEUI, key string) error {
	return s.client.HSet(makeKey(eui), "key", key).Err()
}

func (s *redisAppStorage) Get(eui types.AppEUI) (*App, error) {
	m, err := s.client.HGetAllMap(makeKey(eui)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	app := &App{
		EUI: eui,
		Key: m["key"],
	}
	return app, nil
}

func (s *redisAppStorage) GetAll() ([]*App, error) {
	euis, err := s.client.SMembers(appsKey).Result()
	if err != nil {
		return nil, err
	}
	apps := make([]*App, len(euis))
	for i, k := range euis {
		eui, err := types.ParseAppEUI(k)
		if err != nil {
			return nil, err
		}
		app, err := s.Get(eui)
		if err != nil {
			return nil, err
		}
		apps[i] = app
	}
	return apps, nil
}

func (s *redisAppStorage) Reset() error {
	return s.client.FlushDb().Err()
}

func (s *redisAppStorage) Close() error {
	return s.client.Close()
}
