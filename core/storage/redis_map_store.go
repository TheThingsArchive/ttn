// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"sort"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"

	redis "gopkg.in/redis.v4"
)

// RedisMapStore stores structs as HMaps in Redis
type RedisMapStore struct {
	prefix  string
	client  *redis.Client
	encoder StringStringMapEncoder
	decoder StringStringMapDecoder
}

// NewRedisMapStore returns a new RedisMapStore that talks to the given Redis client and respects the given prefix
func NewRedisMapStore(client *redis.Client, prefix string) *RedisMapStore {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return &RedisMapStore{
		client: client,
		prefix: prefix,
	}
}

// SetBase sets the base struct for automatically encoding and decoding to and from Redis format
func (s *RedisMapStore) SetBase(base interface{}) {
	s.SetEncoder(defaultStructEncoder)
	s.SetDecoder(buildDefaultStructDecoder(base))
}

// SetEncoder sets the encoder to convert structs to Redis format
func (s *RedisMapStore) SetEncoder(encoder StringStringMapEncoder) {
	s.encoder = encoder
}

// SetDecoder sets the decoder to convert structs from Redis format
func (s *RedisMapStore) SetDecoder(decoder StringStringMapDecoder) {
	s.decoder = decoder
}

// GetAll returns all results for the given keys, prepending the prefix to the keys if necessary
func (s *RedisMapStore) GetAll(keys []string, options *ListOptions) ([]interface{}, error) {
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
	cmds := make(map[string]*redis.StringStringMapCmd)
	for _, key := range selectedKeys {
		cmds[key] = pipe.HGetAll(key)
	}

	// Execute pipeline
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	// Get all results from pipeline
	results := make([]interface{}, 0, len(selectedKeys))
	for _, key := range selectedKeys {
		if result, err := cmds[key].Result(); err == nil {
			if result, err := s.decoder(result); err == nil {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// List all results matching the selector, prepending the prefix to the selector if necessary
func (s *RedisMapStore) List(selector string, options *ListOptions) ([]interface{}, error) {
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
func (s *RedisMapStore) Get(key string) (interface{}, error) {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	result, err := s.client.HGetAll(key).Result()
	if err == redis.Nil || len(result) == 0 {
		return nil, errors.NewErrNotFound(key)
	}
	if err != nil {
		return nil, err
	}
	i, err := s.decoder(result)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// GetFields for a record, prepending the prefix to the key if necessary
func (s *RedisMapStore) GetFields(key string, fields ...string) (interface{}, error) {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	result, err := s.client.HMGet(key, fields...).Result()
	if err == redis.Nil {
		return nil, errors.NewErrNotFound(key)
	}
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for i, field := range fields {
		if str, ok := result[i].(string); ok {
			res[field] = str
		}
	}
	i, err := s.decoder(res)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Create a new record, prepending the prefix to the key if necessary, optionally setting only the given properties
func (s *RedisMapStore) Create(key string, value interface{}, properties ...string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}

	vmap, err := s.encoder(value, properties...)
	if err != nil {
		return err
	}
	if len(vmap) == 0 {
		return nil
	}

	err = s.client.Watch(func(tx *redis.Tx) error {
		exists, err := tx.Exists(key).Result()
		if err != nil {
			return err
		}
		if exists {
			return errors.NewErrAlreadyExists(key)
		}
		_, err = tx.MultiExec(func() error {
			tx.HMSet(key, vmap)
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

// Update an existing record, prepending the prefix to the key if necessary, optionally setting only the given properties
func (s *RedisMapStore) Update(key string, value interface{}, properties ...string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}

	vmap, err := s.encoder(value, properties...)
	if err != nil {
		return err
	}
	if len(vmap) == 0 {
		return nil
	}

	err = s.client.Watch(func(tx *redis.Tx) error {
		exists, err := tx.Exists(key).Result()
		if err != nil {
			return err
		}
		if !exists {
			return errors.NewErrNotFound(key)
		}
		_, err = tx.MultiExec(func() error {
			tx.HMSet(key, vmap)
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
func (s *RedisMapStore) Delete(key string) error {
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
