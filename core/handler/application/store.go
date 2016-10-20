// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v4"
)

// Store interface for Applications
type Store interface {
	List() ([]*Application, error)
	Get(appID string) (*Application, error)
	Set(new *Application, properties ...string) (err error)
	Delete(appID string) error
}

const defaultRedisPrefix = "handler"
const redisApplicationPrefix = "application"

// NewRedisApplicationStore creates a new Redis-based Application store
// if an empty prefix is passed, a default prefix will be used.
func NewRedisApplicationStore(client *redis.Client, prefix string) Store {
	if prefix == "" {
		prefix = defaultRedisPrefix
	}
	store := storage.NewRedisMapStore(client, prefix+":"+redisApplicationPrefix)
	store.SetBase(Application{}, "")
	return &RedisApplicationStore{
		store: store,
	}
}

// RedisApplicationStore stores Applications in Redis.
// - Applications are stored as a Hash
type RedisApplicationStore struct {
	store *storage.RedisMapStore
}

// List all Applications
func (s *RedisApplicationStore) List() ([]*Application, error) {
	applicationsI, err := s.store.List("", nil)
	if err != nil {
		return nil, err
	}
	applications := make([]*Application, 0, len(applicationsI))
	for _, applicationI := range applicationsI {
		if application, ok := applicationI.(Application); ok {
			applications = append(applications, &application)
		}
	}
	return applications, nil
}

// Get a specific Application
func (s *RedisApplicationStore) Get(appID string) (*Application, error) {
	applicationI, err := s.store.Get(appID)
	if err != nil {
		return nil, err
	}
	if application, ok := applicationI.(Application); ok {
		return &application, nil
	}
	return nil, errors.New("Database did not return a Application")
}

// Set a new Application or update an existing one
func (s *RedisApplicationStore) Set(new *Application, properties ...string) (err error) {
	now := time.Now()
	new.UpdatedAt = now

	if new.old != nil {
		err = s.store.Update(new.AppID, *new, properties...)
	} else {
		new.CreatedAt = now
		err = s.store.Create(new.AppID, *new, properties...)
	}
	if err != nil {
		return
	}

	return nil
}

// Delete an Application
func (s *RedisApplicationStore) Delete(appID string) error {
	return s.store.Delete(appID)
}
