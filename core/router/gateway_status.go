package router

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	"gopkg.in/redis.v3"
)

type GatewayStatusStore interface {
	Upsert(id types.GatewayEUI, status *gateway.StatusMessage) error
	Get(id types.GatewayEUI) (*gateway.StatusMessage, error)
}

type redisGatewayStore struct {
	client *redis.Client
}

func NewGatewayStatusStore(client *redis.Client) GatewayStatusStore {
	return &redisGatewayStore{
		client: client,
	}
}

func (s *redisGatewayStore) key(gatewayEUI types.GatewayEUI) string {
	return fmt.Sprintf("gateway:%s", gatewayEUI)
}

func (s *redisGatewayStore) Upsert(gatewayEUI types.GatewayEUI, status *gateway.StatusMessage) error {
	m, err := status.ToStringStringMap(gateway.StatusMessageProperties...)
	if err != nil {
		return err
	}
	return s.client.HMSetMap(s.key(gatewayEUI), m).Err()
}

func (s *redisGatewayStore) Get(gatewayEUI types.GatewayEUI) (*gateway.StatusMessage, error) {
	status := &gateway.StatusMessage{}
	res, err := s.client.HGetAllMap(s.key(gatewayEUI)).Result()
	if err != nil {
		return nil, err
	}
	err = status.FromStringStringMap(res)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (r *router) HandleGatewayStatus(gatewayEUI types.GatewayEUI, status *gateway.StatusMessage) error {
	err := r.gatewayStatusStore.Upsert(gatewayEUI, status)
	if err != nil {
		return err
	}
	return nil
}
