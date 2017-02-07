// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"strings"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	redis "gopkg.in/redis.v5"
)

// VersionKey indicates the data (schema) version
const VersionKey = "_version"

type hasDBVersion interface {
	DBVersion() string
}

// MigrateFunction migrates data from its old version to the latest
type MigrateFunction func(client *redis.Client, key string, input map[string]string) (version string, output map[string]string, err error)

// AddMigration adds a data migration for a version
func (s *RedisMapStore) AddMigration(version string, migrate MigrateFunction) {
	s.migrations[version] = migrate
}

// Migrate all documents matching the selector
func (s *RedisMapStore) Migrate(selector string) error {
	if selector == "" {
		selector = "*"
	}
	if !strings.HasPrefix(selector, s.prefix) {
		selector = s.prefix + selector
	}

	var cursor uint64
	for {
		keys, next, err := s.client.Scan(cursor, selector, 0).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			// Get migrates the item
			_, err := s.Get(key)

			// NotFound if item was deleted since Scan started
			if errors.GetErrType(err) == errors.NotFound {
				continue
			}

			// TxFailedError is returned if the item was changed by a concurrent process.
			// If it was a parallel migration, we don't have to do anything
			// Otherwise, this item will be migrated with the next Get
			if err == redis.TxFailedErr {
				continue
			}

			// Any other error should be returned
			if err != nil {
				return err
			}
		}

		cursor = next
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (s *RedisMapStore) migrate(key string, obj map[string]string) (map[string]string, error) {
	var err error

	defer func() {
		if err != nil {
			log.Get().WithField("Key", key).WithError(err).Warn("Data migration failed")
		}
	}()

	version, _ := obj[VersionKey]
	migration, ok := s.migrations[version]

	if !ok {
		return obj, nil
	}

	var oldFields []string
	for k := range obj {
		oldFields = append(oldFields, k)
	}

	err = s.client.Watch(func(tx *redis.Tx) error { // Make sure objects are not changed while we're migrating them
		// As long as there are migrations to do
		for ok {
			version, obj, err = migration(s.client, key, obj)
			if err != nil {
				return err
			}
			obj[VersionKey] = version
			migration, ok = s.migrations[version]
		}

		// Collect deleted fields
		var deletedFields []string
		for _, k := range oldFields {
			if _, ok := obj[k]; !ok {
				deletedFields = append(deletedFields, k)
			}
		}

		// Commit the new version
		_, err := tx.Pipelined(func(pipe *redis.Pipeline) error {
			pipe.HMSet(key, obj)
			if len(deletedFields) > 0 {
				pipe.HDel(key, deletedFields...)
			}
			return nil
		})

		return err
	}, key)

	return obj, err
}
