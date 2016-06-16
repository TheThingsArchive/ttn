package application

import (
	"errors"
	"fmt"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
)

var (
	// ErrNotFound is returned when an application was not found
	ErrNotFound = errors.New("ttn/broker: Application not found")
)

// Store is used to store application configurations
type Store interface {
	// Get the full information about an application
	Get(appEUI types.AppEUI) (*Application, error)
	// Set the given fields of an application. If fields empty, it sets all fields.
	Set(application *Application, fields ...string) error
	// Delete an application
	Delete(types.AppEUI) error
}

// NewApplicationStore creates a new in-memory Application store
func NewApplicationStore() Store {
	return &applicationStore{
		applications: make(map[types.AppEUI]*Application),
	}
}

// applicationStore is an in-memory Application store. It should only be used for testing
// purposes. Use the redisApplicationStore for actual deployments.
type applicationStore struct {
	applications map[types.AppEUI]*Application
}

func (s *applicationStore) Get(appEUI types.AppEUI) (*Application, error) {
	if app, ok := s.applications[appEUI]; ok {
		return app, nil
	}
	return nil, ErrNotFound
}

func (s *applicationStore) Set(new *Application, fields ...string) error {
	// NOTE: We don't care about fields for testing
	s.applications[new.AppEUI] = new
	return nil
}

func (s *applicationStore) Delete(appEUI types.AppEUI) error {
	delete(s.applications, appEUI)
	return nil
}

// NewRedisApplicationStore creates a new Redis-based status store
func NewRedisApplicationStore(client *redis.Client) Store {
	return &redisApplicationStore{
		client: client,
	}
}

const redisApplicationPrefix = "broker:application"

type redisApplicationStore struct {
	client *redis.Client
}

func (s *redisApplicationStore) Get(appEUI types.AppEUI) (*Application, error) {
	res, err := s.client.HGetAllMap(fmt.Sprintf("%s:%s", redisApplicationPrefix, appEUI)).Result()
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
	key := fmt.Sprintf("%s:%s", redisApplicationPrefix, new.AppEUI)
	dmap, err := new.ToStringStringMap(fields...)
	if err != nil {
		return err
	}
	s.client.HMSetMap(key, dmap)
	return nil
}

func (s *redisApplicationStore) Delete(appEUI types.AppEUI) error {
	key := fmt.Sprintf("%s:%s", redisApplicationPrefix, appEUI)
	err := s.client.Del(key).Err()
	if err != nil {
		return err
	}
	return nil
}
