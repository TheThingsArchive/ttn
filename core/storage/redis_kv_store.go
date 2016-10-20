// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"sort"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"

	"gopkg.in/redis.v4"
)

// RedisKVStore stores arbitrary data in Redis
type RedisKVStore struct {
	prefix string
	client *redis.Client
}

// NewRedisKVStore creates a new RedisKVStore
func NewRedisKVStore(client *redis.Client, prefix string) *RedisKVStore {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return &RedisKVStore{
		client: client,
		prefix: prefix,
	}
}

// GetAll returns all results for the given keys, prepending the prefix to the keys if necessary
func (s *RedisKVStore) GetAll(keys []string, options *ListOptions) (map[string]string, error) {
	for i, key := range keys {
		if !strings.HasPrefix(key, s.prefix) {
			keys[i] = s.prefix + key
		}
	}

	sort.Strings(keys)

	selectedKeys := selectKeys(keys, options)

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
	if selector == "" {
		selector = "*"
	}
	if !strings.HasPrefix(selector, s.prefix) {
		selector = s.prefix + selector
	}
	keys, err := s.client.Keys(selector).Result()
	if err != nil {
		return nil, err
	}
	return s.GetAll(keys, options)
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

// Create a new record, prepending the prefix to the key if necessary
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
		_, err = tx.MultiExec(func() error {
			tx.Set(key, value, 0)
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
		_, err = tx.MultiExec(func() error {
			tx.Set(key, value, 0)
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

// Delete an existing record, prepending the prefix to the key if necessary
func (s *RedisKVStore) Delete(key string) error {
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
		_, err = tx.MultiExec(func() error {
			tx.Del(key)
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
