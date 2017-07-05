// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/core/discovery/announcement/migrate"
	"github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"gopkg.in/redis.v5"
)

// Store interface for Announcements
type Store interface {
	List(opts *storage.ListOptions) ([]*Announcement, error)
	ListService(serviceName string, opts *storage.ListOptions) ([]*Announcement, error)
	Get(serviceName, serviceID string) (*Announcement, error)
	GetMetadata(serviceName, serviceID string) ([]Metadata, error)
	getForAppID(appID string) (serviceName, serviceID string, err error)
	GetForAppID(appID string) (*Announcement, error)
	getForAppEUI(appEUI types.AppEUI) (serviceName, serviceID string, err error)
	GetForAppEUI(appEUI types.AppEUI) (*Announcement, error)
	Set(new *Announcement) error
	AddMetadata(serviceName, serviceID string, metadata ...Metadata) error
	RemoveMetadata(serviceName, serviceID string, metadata ...Metadata) error
	Delete(serviceName, serviceID string) error
}

const defaultRedisPrefix = "discovery"

const redisAnnouncementPrefix = "announcement"
const redisMetadataPrefix = "metadata"
const redisAppIDPrefix = "app_id"
const redisAppEUIPrefix = "app_eui"

// NewRedisAnnouncementStore creates a new Redis-based Announcement store
func NewRedisAnnouncementStore(client *redis.Client, prefix string) Store {
	if prefix == "" {
		prefix = defaultRedisPrefix
	}
	store := storage.NewRedisMapStore(client, prefix+":"+redisAnnouncementPrefix)
	store.SetBase(Announcement{}, "")
	for v, f := range migrate.AnnouncementMigrations(prefix) {
		store.AddMigration(v, f)
	}
	return &RedisAnnouncementStore{
		store:    store,
		metadata: storage.NewRedisSetStore(client, prefix+":"+redisMetadataPrefix),
		byAppID:  storage.NewRedisKVStore(client, prefix+":"+redisAppIDPrefix),
		byAppEUI: storage.NewRedisKVStore(client, prefix+":"+redisAppEUIPrefix),
	}
}

// RedisAnnouncementStore stores Announcements in Redis.
// - Announcements are stored as a Hash
// - Metadata is stored in a Set
// - AppIDs and AppEUIs are indexed with key/value pairs
type RedisAnnouncementStore struct {
	store    *storage.RedisMapStore
	metadata *storage.RedisSetStore
	byAppID  *storage.RedisKVStore
	byAppEUI *storage.RedisKVStore
}

// List all Announcements
// The resulting Announcements do *not* include metadata
func (s *RedisAnnouncementStore) List(opts *storage.ListOptions) ([]*Announcement, error) {
	announcementsI, err := s.store.List("", opts)
	if err != nil {
		return nil, err
	}
	announcements := make([]*Announcement, len(announcementsI))
	for i, announcementI := range announcementsI {
		if announcement, ok := announcementI.(Announcement); ok {
			announcements[i] = &announcement
		}
	}
	return announcements, nil
}

// ListService lists all Announcements for a given service (router/broker/handler)
// The resulting Announcements *do* include metadata
func (s *RedisAnnouncementStore) ListService(serviceName string, opts *storage.ListOptions) ([]*Announcement, error) {
	announcementsI, err := s.store.List(serviceName+":*", opts)
	if err != nil {
		return nil, err
	}
	announcements := make([]*Announcement, len(announcementsI))
	for i, announcementI := range announcementsI {
		if announcement, ok := announcementI.(Announcement); ok {
			announcements[i] = &announcement
			announcement.Metadata, err = s.GetMetadata(announcement.ServiceName, announcement.ID)
			if err != nil {
				return nil, err
			}
		}
	}
	return announcements, nil
}

// Get a specific service Announcement
// The result *does* include metadata
func (s *RedisAnnouncementStore) Get(serviceName, serviceID string) (*Announcement, error) {
	announcementI, err := s.store.Get(fmt.Sprintf("%s:%s", serviceName, serviceID))
	if err != nil {
		return nil, err
	}
	announcement, ok := announcementI.(Announcement)
	if !ok {
		return nil, errors.New("Database did not return an Announcement")
	}
	announcement.Metadata, err = s.GetMetadata(serviceName, serviceID)
	if err != nil {
		return nil, err
	}
	return &announcement, nil
}

// GetMetadata returns the metadata of the specified service
func (s *RedisAnnouncementStore) GetMetadata(serviceName, serviceID string) ([]Metadata, error) {
	var out []Metadata
	metadata, err := s.metadata.Get(fmt.Sprintf("%s:%s", serviceName, serviceID))
	if errors.GetErrType(err) == errors.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for _, meta := range metadata {
		if meta := MetadataFromString(meta); meta != nil {
			out = append(out, meta)
		}
	}
	return out, nil
}

func (s *RedisAnnouncementStore) getForAppID(appID string) (string, string, error) {
	key, err := s.byAppID.Get(appID)
	if err != nil {
		return "", "", err
	}
	service := strings.Split(key, ":")
	return service[0], service[1], nil
}

// GetForAppID returns the last Announcement that contains metadata for the given AppID
func (s *RedisAnnouncementStore) GetForAppID(appID string) (*Announcement, error) {
	serviceName, serviceID, err := s.getForAppID(appID)
	if err != nil {
		return nil, err
	}
	return s.Get(serviceName, serviceID)
}

func (s *RedisAnnouncementStore) getForAppEUI(appEUI types.AppEUI) (string, string, error) {
	key, err := s.byAppEUI.Get(appEUI.String())
	if err != nil {
		return "", "", err
	}
	service := strings.Split(key, ":")
	return service[0], service[1], nil
}

// GetForAppEUI returns the last Announcement that contains metadata for the given AppEUI
func (s *RedisAnnouncementStore) GetForAppEUI(appEUI types.AppEUI) (*Announcement, error) {
	serviceName, serviceID, err := s.getForAppEUI(appEUI)
	if err != nil {
		return nil, err
	}
	return s.Get(serviceName, serviceID)
}

// Set a new Announcement or update an existing one
// The metadata of the announcement is ignored, as metadata should be managed with AddMetadata and RemoveMetadata
func (s *RedisAnnouncementStore) Set(new *Announcement) error {
	now := time.Now()
	new.UpdatedAt = now
	key := fmt.Sprintf("%s:%s", new.ServiceName, new.ID)
	if new.old == nil {
		new.CreatedAt = now
	}
	err := s.store.Set(key, *new)
	if err != nil {
		return err
	}
	return nil
}

// AddMetadata adds metadata to the announcement of the specified service
func (s *RedisAnnouncementStore) AddMetadata(serviceName, serviceID string, metadata ...Metadata) error {
	key := fmt.Sprintf("%s:%s", serviceName, serviceID)

	metadataStrings := make([]string, 0, len(metadata))
	for _, meta := range metadata {
		txt, err := meta.MarshalText()
		if err != nil {
			return err
		}
		metadataStrings = append(metadataStrings, string(txt))

		switch meta := meta.(type) {
		case AppIDMetadata:
			existing, err := s.byAppID.Get(meta.AppID)
			switch {
			case errors.GetErrType(err) == errors.NotFound:
				if err := s.byAppID.Create(meta.AppID, key); err != nil {
					return err
				}
			case err != nil:
				return err
			case existing == key:
				continue
			default:
				go s.metadata.Remove(existing, string(txt))
				if err := s.byAppID.Update(meta.AppID, key); err != nil {
					return err
				}
			}
		case AppEUIMetadata:
			existing, err := s.byAppEUI.Get(meta.AppEUI.String())
			switch {
			case errors.GetErrType(err) == errors.NotFound:
				if err := s.byAppEUI.Create(meta.AppEUI.String(), key); err != nil {
					return err
				}
			case err != nil:
				return err
			case existing == key:
				continue
			default:
				go s.metadata.Remove(existing, string(txt))
				if err := s.byAppEUI.Update(meta.AppEUI.String(), key); err != nil {
					return err
				}
			}
		}
	}
	err := s.metadata.Add(key, metadataStrings...)
	if err != nil {
		return err
	}
	return nil
}

// RemoveMetadata removes metadata from the announcement of the specified service
func (s *RedisAnnouncementStore) RemoveMetadata(serviceName, serviceID string, metadata ...Metadata) error {
	metadataStrings := make([]string, 0, len(metadata))
	for _, meta := range metadata {
		if txt, err := meta.MarshalText(); err == nil {
			metadataStrings = append(metadataStrings, string(txt))
		}
		switch meta := meta.(type) {
		case AppIDMetadata:
			s.byAppID.Delete(meta.AppID)
		case AppEUIMetadata:
			s.byAppEUI.Delete(meta.AppEUI.String())
		}
	}
	err := s.metadata.Remove(fmt.Sprintf("%s:%s", serviceName, serviceID), metadataStrings...)
	if err != nil {
		return err
	}
	return nil
}

// Delete an Announcement and its metadata
func (s *RedisAnnouncementStore) Delete(serviceName, serviceID string) error {
	metadata, err := s.GetMetadata(serviceName, serviceID)
	if err != nil && errors.GetErrType(err) != errors.NotFound {
		return err
	}
	if len(metadata) > 0 {
		s.RemoveMetadata(serviceName, serviceID, metadata...)
	}
	key := fmt.Sprintf("%s:%s", serviceName, serviceID)
	err = s.store.Delete(key)
	if err != nil {
		return err
	}
	return nil
}
