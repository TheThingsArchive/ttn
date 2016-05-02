// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/collection"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// App represents a stored application
type App struct {
	EUI types.AppEUI
	Key string
	collection.Functions
}

// AppStorage provides storage for applications
type AppStorage interface {
	SetKey(eui types.AppEUI, key string) error
	SetFunctions(eui types.AppEUI, functions *collection.Functions) error
	Get(eui types.AppEUI) (*App, error)
	GetAll() ([]*App, error)
	Reset() error
	Close() error
}

type redisAppStorage struct {
	client *redis.Client
}

// ConnectRedisAppStorage connects to Redis using the specified options
func ConnectRedisAppStorage(opt *redis.Options) (AppStorage, error) {
	client := redis.NewClient(opt)
	_, err := client.Ping().Result()
	if err != nil {
		client.Close()
		return nil, err
	}
	return &redisAppStorage{client}, nil
}

func (s *redisAppStorage) SetKey(eui types.AppEUI, key string) error {
	return s.client.HSet(eui.String(), "key", key).Err()
}

func (s *redisAppStorage) SetFunctions(eui types.AppEUI, functions *collection.Functions) error {
	return s.client.HMSetMap(eui.String(), map[string]string{
		"decoder":   functions.Decoder,
		"converter": functions.Converter,
		"validator": functions.Validator,
	}).Err()
}

func (s *redisAppStorage) Get(eui types.AppEUI) (*App, error) {
	m, err := s.client.HGetAllMap(eui.String()).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	app := &App{
		EUI: eui,
		Key: m["key"],
		Functions: collection.Functions{
			Decoder:   m["decoder"],
			Converter: m["converter"],
			Validator: m["validator"],
		},
	}
	return app, nil
}

func (s *redisAppStorage) GetAll() ([]*App, error) {
	euis, err := s.client.Keys("*").Result()
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
