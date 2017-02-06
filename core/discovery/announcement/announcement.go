// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/fatih/structs"
)

const currentDBVersion = "2.4.1"

// Metadata represents metadata that is stored with an Announcement
type Metadata interface {
	encoding.TextMarshaler
	ToProto() *pb.Metadata
}

// AppEUIMetadata is used to store an AppEUI
type AppEUIMetadata struct {
	AppEUI types.AppEUI
}

// ToProto implements the Metadata interface
func (m AppEUIMetadata) ToProto() *pb.Metadata {
	return &pb.Metadata{
		Metadata: &pb.Metadata_AppEui{
			AppEui: m.AppEUI.Bytes(),
		},
	}
}

// MarshalText implements the encoding.TextMarshaler interface
func (m AppEUIMetadata) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("AppEUI %s", m.AppEUI)), nil
}

// AppIDMetadata is used to store an AppID
type AppIDMetadata struct {
	AppID string
}

// ToProto implements the Metadata interface
func (m AppIDMetadata) ToProto() *pb.Metadata {
	return &pb.Metadata{
		Metadata: &pb.Metadata_AppId{
			AppId: m.AppID,
		},
	}
}

// MarshalText implements the encoding.TextMarshaler interface
func (m AppIDMetadata) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("AppID %s", m.AppID)), nil
}

// PrefixMetadata is used to store a DevAddr prefix
type PrefixMetadata struct {
	Prefix types.DevAddrPrefix
}

// ToProto implements the Metadata interface
func (m PrefixMetadata) ToProto() *pb.Metadata {
	return &pb.Metadata{
		Metadata: &pb.Metadata_DevAddrPrefix{
			DevAddrPrefix: m.Prefix.Bytes(),
		},
	}
}

// MarshalText implements the encoding.TextMarshaler interface
func (m PrefixMetadata) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("Prefix %s", m.Prefix)), nil
}

// MetadataFromProto converts a protocol buffer metadata to a Metadata
func MetadataFromProto(proto *pb.Metadata) Metadata {
	if euiBytes := proto.GetAppEui(); euiBytes != nil {
		eui := new(types.AppEUI)
		if err := eui.Unmarshal(euiBytes); err != nil {
			return nil
		}
		return AppEUIMetadata{*eui}
	}
	if id := proto.GetAppId(); id != "" {
		return AppIDMetadata{id}
	}
	if prefixBytes := proto.GetDevAddrPrefix(); prefixBytes != nil {
		prefix := new(types.DevAddrPrefix)
		if err := prefix.Unmarshal(prefixBytes); err != nil {
			return nil
		}
		return PrefixMetadata{*prefix}
	}
	return nil
}

// MetadataFromString converts a string to a Metadata
func MetadataFromString(str string) Metadata {
	meta := strings.SplitAfterN(str, " ", 2)
	key := strings.TrimSpace(meta[0])
	value := meta[1]
	switch key {
	case "AppEUI":
		var appEUI types.AppEUI
		appEUI.UnmarshalText([]byte(value))
		return AppEUIMetadata{appEUI}
	case "AppID":
		return AppIDMetadata{value}
	case "Prefix":
		prefix := &types.DevAddrPrefix{
			Length: 32,
		}
		prefix.UnmarshalText([]byte(value))
		return PrefixMetadata{*prefix}
	}
	return nil
}

// Announcement of a network component
type Announcement struct {
	old *Announcement

	ID             string `redis:"id"`
	ServiceName    string `redis:"service_name"`
	ServiceVersion string `redis:"service_version"`
	Description    string `redis:"description"`
	URL            string `redis:"url"`
	Public         bool   `redis:"public"`
	NetAddress     string `redis:"net_address"`
	PublicKey      string `redis:"public_key"`
	Certificate    string `redis:"certificate"`
	APIAddress     string `redis:"api_address"`
	MQTTAddress    string `redis:"mqtt_address"`
	AMQPAddress    string `redis:"amqp_address"`
	Metadata       []Metadata

	CreatedAt time.Time `redis:"created_at"`
	UpdatedAt time.Time `redis:"updated_at"`
}

// StartUpdate stores the state of the announcement
func (a *Announcement) StartUpdate() {
	old := *a
	a.old = &old
}

// DBVersion of the model
func (a *Announcement) DBVersion() string {
	return currentDBVersion
}

// ChangedFields returns the names of the changed fields since the last call to StartUpdate
func (a Announcement) ChangedFields() (changed []string) {
	new := structs.New(a)
	fields := new.Names()
	if a.old == nil {
		return fields
	}
	old := structs.New(*a.old)

	for _, field := range new.Fields() {
		if !field.IsExported() || field.Name() == "old" {
			continue
		}
		if !reflect.DeepEqual(field.Value(), old.Field(field.Name()).Value()) {
			changed = append(changed, field.Name())
		}
	}
	return
}

// ToProto converts the Announcement to a protobuf Announcement
func (a Announcement) ToProto() *pb.Announcement {
	metadata := make([]*pb.Metadata, 0, len(a.Metadata))
	for _, meta := range a.Metadata {
		metadata = append(metadata, meta.ToProto())
	}
	return &pb.Announcement{
		Id:             a.ID,
		ServiceName:    a.ServiceName,
		ServiceVersion: a.ServiceVersion,
		Description:    a.Description,
		Url:            a.URL,
		Public:         a.Public,
		NetAddress:     a.NetAddress,
		PublicKey:      a.PublicKey,
		Certificate:    a.Certificate,
		ApiAddress:     a.APIAddress,
		MqttAddress:    a.MQTTAddress,
		AmqpAddress:    a.AMQPAddress,
		Metadata:       metadata,
	}
}

// FromProto converts an Announcement protobuf to an Announcement
func FromProto(a *pb.Announcement) Announcement {
	metadata := make([]Metadata, 0, len(a.Metadata))
	for _, meta := range a.Metadata {
		metadata = append(metadata, MetadataFromProto(meta))
	}
	return Announcement{
		ID:             a.Id,
		ServiceName:    a.ServiceName,
		ServiceVersion: a.ServiceVersion,
		Description:    a.Description,
		URL:            a.Url,
		Public:         a.Public,
		NetAddress:     a.NetAddress,
		PublicKey:      a.PublicKey,
		Certificate:    a.Certificate,
		APIAddress:     a.ApiAddress,
		MQTTAddress:    a.MqttAddress,
		AMQPAddress:    a.AmqpAddress,
		Metadata:       metadata,
	}
}
