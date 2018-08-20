// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"strings"

	"gopkg.in/redis.v5"
)

// RedisStore is the base of more specialized stores
type RedisStore struct {
	prefix string
	client *redis.Client
}

// NewRedisStore creates a new RedisStore
func NewRedisStore(client *redis.Client, prefix string) *RedisStore {
	if !strings.HasSuffix(prefix, ":") {
		prefix += ":"
	}
	return &RedisStore{
		client: client,
		prefix: prefix,
	}
}

// Keys matching the selector, prepending the prefix to the selector if necessary
func (s *RedisStore) Keys(selector string, opts *ListOptions) ([]string, error) {
	if opts != nil && opts.cacheKey != "" {
		cacheExists, err := s.client.Exists(opts.cacheKey).Result()
		if err != nil {
			return nil, err
		}
		if cacheExists {
			res, err := s.client.SMembers(opts.cacheKey).Result()
			if err == nil {
				return res, nil
			}
		}
	}

	if selector == "" {
		selector = "*"
	}
	if !strings.HasPrefix(selector, s.prefix) {
		selector = s.prefix + selector
	}
	var allKeys []string
	var cursor uint64
	for {
		keys, next, err := s.client.Scan(cursor, selector, 10000).Result()
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		cursor = next
		if cursor == 0 {
			break
		}
	}

	if opts != nil && opts.cacheKey != "" {
		pipe := s.client.Pipeline()
		pipe.Del(opts.cacheKey)
		var allKeysAsInterface = make([]interface{}, len(allKeys))
		for i, key := range allKeys {
			allKeysAsInterface[i] = key
		}
		pipe.SAdd(opts.cacheKey, allKeysAsInterface...)
		if opts.cacheTTL > 0 {
			pipe.Expire(opts.cacheKey, opts.cacheTTL)
		}
	}

	return allKeys, nil
}

// Count the results matching the selector
func (s *RedisStore) Count(selector string, options *ListOptions) (int, error) {
	allKeys, err := s.Keys(selector, options)
	if err != nil {
		return 0, err
	}
	return len(allKeys), nil
}

// Delete an existing record, prepending the prefix to the key if necessary
func (s *RedisStore) Delete(key string) error {
	if !strings.HasPrefix(key, s.prefix) {
		key = s.prefix + key
	}
	return s.client.Del(key).Err()
}
