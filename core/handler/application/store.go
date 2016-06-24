// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"errors"
	"fmt"

	"gopkg.in/redis.v3"
)

var (
	// ErrNotFound is returned when a application was not found
	ErrNotFound = errors.New("ttn/handler: Application not found")
)

// Store is used to store application configurations
type Store interface {
	// List all applications
	List() ([]*Application, error)
	// Get the full information about a application
	Get(appID string) (*Application, error)
	// Set the given fields of a application. If fields empty, it sets all fields.
	Set(application *Application, fields ...string) error
	// Delete a application
	Delete(appid string) error
}

// NewApplicationStore creates a new in-memory Application store
func NewApplicationStore() Store {
	return &applicationStore{
		applications: make(map[string]*Application),
	}
}

// applicationStore is an in-memory Application store. It should only be used for testing
// purposes. Use the redisApplicationStore for actual deployments.
type applicationStore struct {
	applications map[string]*Application
}

func (s *applicationStore) List() ([]*Application, error) {
	apps := make([]*Application, 0, len(s.applications))
	for _, application := range s.applications {
		apps = append(apps, application)
	}
	return apps, nil
}

func (s *applicationStore) Get(appID string) (*Application, error) {
	if app, ok := s.applications[appID]; ok {
		return app, nil
	}
	return nil, ErrNotFound
}

func (s *applicationStore) Set(new *Application, fields ...string) error {
	// NOTE: We don't care about fields for testing
	s.applications[new.AppID] = new
	return nil
}

func (s *applicationStore) Delete(appID string) error {
	delete(s.applications, appID)
	return nil
}

// NewRedisApplicationStore creates a new Redis-based status store
func NewRedisApplicationStore(client *redis.Client) Store {
	return &redisApplicationStore{
		client: client,
	}
}

const redisApplicationPrefix = "handler:application"

type redisApplicationStore struct {
	client *redis.Client
}

func (s *redisApplicationStore) List() ([]*Application, error) {
	var apps []*Application
	keys, err := s.client.Keys(fmt.Sprintf("%s:*", redisApplicationPrefix)).Result()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		res, err := s.client.HGetAllMap(key).Result()
		if err != nil {
			return nil, err
		}
		application := &Application{}
		err = application.FromStringStringMap(res)
		if err != nil {
			return nil, err
		}
		apps = append(apps, application)
	}
	return apps, nil
}

func (s *redisApplicationStore) Get(appID string) (*Application, error) {
	res, err := s.client.HGetAllMap(fmt.Sprintf("%s:%s", redisApplicationPrefix, appID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	} else if len(res) == 0 {
		return nil, ErrNotFound
	}
	application := &Application{}
	err = application.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func (s *redisApplicationStore) Set(new *Application, fields ...string) error {
	if len(fields) == 0 {
		fields = ApplicationProperties
	}

	key := fmt.Sprintf("%s:%s", redisApplicationPrefix, new.AppID)
	dmap, err := new.ToStringStringMap(fields...)
	if err != nil {
		return err
	}
	err = s.client.HMSetMap(key, dmap).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *redisApplicationStore) Delete(appID string) error {
	key := fmt.Sprintf("%s:%s", redisApplicationPrefix, appID)
	err := s.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}
