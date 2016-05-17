// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"fmt"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
)

const appKey = "fields:app:%s"

// AppStorage provides storage for applications
type AppStorage interface {
	SetFunctions(eui types.AppEUI, functions *Functions) error
	GetFunctions(eui types.AppEUI) (*Functions, error)
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

func (s *redisAppStorage) SetFunctions(eui types.AppEUI, functions *Functions) error {
	return s.client.HMSetMap(makeKey(eui), map[string]string{
		"decoder":   functions.Decoder,
		"converter": functions.Converter,
		"validator": functions.Validator,
	}).Err()
}

func (s *redisAppStorage) GetFunctions(eui types.AppEUI) (*Functions, error) {
	m, err := s.client.HGetAllMap(makeKey(eui)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	decoder, ok := m["decoder"]
	if !ok {
		return nil, nil
	}
	return &Functions{
		Decoder:   decoder,
		Converter: m["converter"],
		Validator: m["validator"],
	}, nil
}

func (s *redisAppStorage) Reset() error {
	return s.client.FlushDb().Err()
}

func (s *redisAppStorage) Close() error {
	return s.client.Close()
}
