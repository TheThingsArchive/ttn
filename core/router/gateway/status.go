package gateway

import (
	"fmt"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	"gopkg.in/redis.v3"
)

// StatusStore is a database for setting and retrieving the latest gateway status
type StatusStore interface {
	// Insert or Update the status
	Update(status *pb_gateway.StatusMessage) error
	// Get the last status
	Get() (*pb_gateway.StatusMessage, error)
}

// NewStatusStore creates a new in-memory status store
func NewStatusStore() StatusStore {
	return &statusStore{}
}

type statusStore struct {
	lastStatus *pb_gateway.StatusMessage
}

func (s *statusStore) Update(status *pb_gateway.StatusMessage) error {
	s.lastStatus = status
	return nil
}

func (s *statusStore) Get() (*pb_gateway.StatusMessage, error) {
	if s.lastStatus != nil {
		return s.lastStatus, nil
	}
	return &pb_gateway.StatusMessage{}, nil
}

// NewRedisStatusStore creates a new Redis-based status store
func NewRedisStatusStore(client *redis.Client, eui types.GatewayEUI) StatusStore {
	return &redisStatusStore{
		client: client,
		key:    fmt.Sprintf("gateway:%s", eui),
	}
}

type redisStatusStore struct {
	client *redis.Client
	key    string
}

func (s *redisStatusStore) Update(status *pb_gateway.StatusMessage) error {
	m, err := status.ToStringStringMap(pb_gateway.StatusMessageProperties...)
	if err != nil {
		return err
	}
	return s.client.HMSetMap(s.key, m).Err()
}

func (s *redisStatusStore) Get() (*pb_gateway.StatusMessage, error) {
	status := &pb_gateway.StatusMessage{}
	res, err := s.client.HGetAllMap(s.key).Result()
	if err != nil {
		return nil, err
	}
	err = status.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	return status, nil
}
