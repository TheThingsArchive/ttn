// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"sort"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v5"
)

// RedisKVStore stores arbitrary data in Redis
type RedisKVStore struct {
	*RedisStore
}

// NewRedisKVStore creates a new RedisKVStore
func NewRedisKVStore(client *redis.Client, prefix string) *RedisKVStore {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return &RedisKVStore{
		RedisStore: NewRedisStore(client, prefix),
	}
}

// GetAll returns all results for the given keys, prepending the prefix to the keys if necessary
func (s *RedisKVStore) GetAll(keys []string, options *ListOptions) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	for i, key := range keys {
		if !strings.HasPrefix(key, s.prefix) {
			keys[i] = s.prefix + key
		}
	}

	sort.Strings(keys)

	selectedKeys := selectKeys(keys, options)
	if len(selectedKeys) == 0 {
		return map[string]string{}, nil
	}

	pipe := s.client.Pipeline()
	defer pipe.Close()

	// Add all commands to pipeline
	cmds := make(map[string]*redis.StringCmd)
	for _, key := range selectedKeys {
		cmds[key] = pipe.Get(key)
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
			data[strings.TrimPrefix(key, s.prefix)] = res
		}
	}

	return data, nil
}

// List all results matching the selector, prepending the prefix to the selector if necessary
func (s *RedisKVStore) List(selector string, options *ListOptions) (map[string]string, error) {
	allKeys, err := s.Keys(selector)
	if err != nil {
		return nil, err
	}
	return s.GetAll(allKeys, options)
}

// Get one result, prepending the prefix to the key if necessary
func (s *RedisKVStore) Get(key string) (string, error) {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	result, err := s.client.Get(key).Result()
	if err == redis.Nil || result == "" {
		return "", errors.NewErrNotFound(key)
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

// Set a record, prepending the prefix to the key if necessary
func (s *RedisKVStore) Set(key string, value string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	return s.client.Set(key, value, 0).Err()
}

// Create a new record, prepending the prefix to the key if necessary
// This function returns an error if the record already exists
func (s *RedisKVStore) Create(key string, value string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	err := s.client.Watch(func(tx *redis.Tx) error {
		exists, err := tx.Exists(key).Result()
		if err != nil {
			return err
		}
		if exists {
			return errors.NewErrAlreadyExists(key)
		}
		_, err = tx.Pipelined(func(pipe *redis.Pipeline) error {
			pipe.Set(key, value, 0)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	}, key)
	if err != nil {
		return err
	}
	return nil
}

// Update an existing record, prepending the prefix to the key if necessary
// This function returns an error if the record does not exist
func (s *RedisKVStore) Update(key string, value string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	err := s.client.Watch(func(tx *redis.Tx) error {
		exists, err := tx.Exists(key).Result()
		if err != nil {
			return err
		}
		if !exists {
			return errors.NewErrNotFound(key)
		}
		_, err = tx.Pipelined(func(pipe *redis.Pipeline) error {
			pipe.Set(key, value, 0)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	}, key)
	if err != nil {
		return err
	}
	return nil
}
