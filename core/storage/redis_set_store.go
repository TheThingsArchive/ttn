// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"sort"
	"strings"

	"github.com/TheThingsNetwork/ttn/utils/errors"

	"gopkg.in/redis.v5"
)

// RedisSetStore stores sets in Redis
type RedisSetStore struct {
	prefix string
	client *redis.Client
}

// NewRedisSetStore creates a new RedisSetStore
func NewRedisSetStore(client *redis.Client, prefix string) *RedisSetStore {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return &RedisSetStore{
		client: client,
		prefix: prefix,
	}
}

// GetAll returns all results for the given keys, prepending the prefix to the keys if necessary
func (s *RedisSetStore) GetAll(keys []string, options *ListOptions) (map[string][]string, error) {
	if len(keys) == 0 {
		return map[string][]string{}, nil
	}

	for i, key := range keys {
		if !strings.HasPrefix(key, s.prefix) {
			keys[i] = s.prefix + key
		}
	}

	sort.Strings(keys)

	selectedKeys := selectKeys(keys, options)
	if len(selectedKeys) == 0 {
		return map[string][]string{}, nil
	}

	pipe := s.client.Pipeline()
	defer pipe.Close()

	// Add all commands to pipeline
	cmds := make(map[string]*redis.StringSliceCmd)
	for _, key := range selectedKeys {
		cmds[key] = pipe.SMembers(key)
	}

	// Execute pipeline
	_, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	// Get all results from pipeline
	data := make(map[string][]string)
	for key, cmd := range cmds {
		res, err := cmd.Result()
		if err == nil {
			sort.Strings(res)
			data[strings.TrimPrefix(key, s.prefix)] = res
		}
	}

	return data, nil
}

// List all results matching the selector, prepending the prefix to the selector if necessary
func (s *RedisSetStore) List(selector string, options *ListOptions) (map[string][]string, error) {
	if selector == "" {
		selector = "*"
	}
	if !strings.HasPrefix(selector, s.prefix) {
		selector = s.prefix + selector
	}
	var allKeys []string
	var cursor uint64
	for {
		keys, next, err := s.client.Scan(cursor, selector, 0).Result()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return s.GetAll(allKeys, options)
}

// Get one result, prepending the prefix to the key if necessary
func (s *RedisSetStore) Get(key string) (res []string, err error) {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	res, err = s.client.SMembers(key).Result()
	if err == redis.Nil || len(res) == 0 {
		return res, errors.NewErrNotFound(key)
	}
	sort.Strings(res)
	return res, err
}

// Contains returns wheter the set contains a given value, prepending the prefix to the key if necessary
func (s *RedisSetStore) Contains(key string, value string) (res bool, err error) {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	res, err = s.client.SIsMember(key, value).Result()
	if err == redis.Nil {
		return res, errors.NewErrNotFound(key)
	}
	return res, err
}

// Add one or more values to the set, prepending the prefix to the key if necessary
func (s *RedisSetStore) Add(key string, values ...string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	valuesI := make([]interface{}, len(values))
	for i, v := range values {
		valuesI[i] = v
	}
	return s.client.SAdd(key, valuesI...).Err()
}

// Remove one or more values from the set, prepending the prefix to the key if necessary
func (s *RedisSetStore) Remove(key string, values ...string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	valuesI := make([]interface{}, len(values))
	for i, v := range values {
		valuesI[i] = v
	}
	return s.client.SRem(key, valuesI...).Err()
}

// Delete the entire set
func (s *RedisSetStore) Delete(key string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	return s.client.Del(key).Err()
}
