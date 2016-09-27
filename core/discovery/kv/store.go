// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package kv

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/utils/errors"

	"gopkg.in/redis.v3"
)

// Store is a simple String/String Key-Value store
type Store interface {
	// List all items in the store
	List() (map[string]string, error)
	// Get the full HandlerID for the AppID
	Get(key string) (value string, err error)
	// Set the mapping.
	Set(key, value string) error
	// Delete a mapping
	Delete(key string) error
}

// NewKVStore creates a new in-memory Key-Value store
func NewKVStore() Store {
	return &kvStore{
		data: make(map[string]string),
	}
}

// kvStore is an in-memory Key-Value store. It should only be used for testing
// purposes. Use the redisKVStore for actual deployments.
type kvStore struct {
	data map[string]string
}

func (s *kvStore) List() (map[string]string, error) {
	return s.data, nil
}

func (s *kvStore) Get(key string) (string, error) {
	if value, ok := s.data[key]; ok {
		return value, nil
	}
	return "", errors.NewErrNotFound(fmt.Sprintf("Discovery: %s", key))
}

func (s *kvStore) Set(key, value string) error {
	s.data[key] = value
	return nil
}

func (s *kvStore) Delete(key string) error {
	delete(s.data, key)
	return nil
}

const redisKVPrefix = "discovery:"

// NewRedisKVStore creates a new Redis-based Key-Value store
func NewRedisKVStore(client *redis.Client, prefix string) Store {
	return &redisKVStore{
		client: client,
		prefix: redisKVPrefix + prefix,
	}
}

type redisKVStore struct {
	prefix string
	client *redis.Client
}

func (s *redisKVStore) getForKeys(keys []string) (map[string]string, error) {
	pipe := s.client.Pipeline()
	defer pipe.Close()

	// Add all commands to pipeline
	cmds := make(map[string]*redis.StringCmd)
	for _, key := range keys {
		cmds[key] = s.client.Get(key)
	}

	// Execute pipeline
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	// Get all results from pipeline
	data := make(map[string]string)
	for key, cmd := range cmds {
		res, err := cmd.Result()
		if err == nil {
			data[key] = res
		}
	}

	return data, nil
}

func (s *redisKVStore) List() (map[string]string, error) {
	keys, err := s.client.Keys(fmt.Sprintf("%s:*", s.prefix)).Result()
	if err != nil {
		return nil, err
	}
	return s.getForKeys(keys)
}

func (s *redisKVStore) Get(key string) (string, error) {
	res, err := s.client.Get(fmt.Sprintf("%s:%s", s.prefix, key)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.NewErrNotFound(fmt.Sprintf("Discovery: %s", key))
		}
		return "", err
	} else if res == "" {
		return "", errors.NewErrNotFound(fmt.Sprintf("Discovery: %s", key))
	}
	return res, nil
}

func (s *redisKVStore) Set(key, value string) error {
	err := s.client.Set(fmt.Sprintf("%s:%s", s.prefix, key), value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *redisKVStore) Delete(key string) error {
	err := s.client.Del(fmt.Sprintf("%s:%s", s.prefix, key)).Err()
	if err != nil {
		return err
	}
	return nil
}
