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
		Key:   pb.Metadata_APP_EUI,
		Value: m.AppEUI.Bytes(),
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
		Key:   pb.Metadata_APP_ID,
		Value: []byte(m.AppID),
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
		Key:   pb.Metadata_PREFIX,
		Value: []byte{byte(m.Prefix.Length), m.Prefix.DevAddr[0], m.Prefix.DevAddr[1], m.Prefix.DevAddr[2], m.Prefix.DevAddr[3]},
	}
}

// MarshalText implements the encoding.TextMarshaler interface
func (m PrefixMetadata) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("Prefix %s", m.Prefix)), nil
}

// OtherMetadata is used to store arbitrary information
type OtherMetadata struct {
	Data string
}

// ToProto implements the Metadata interface
func (m OtherMetadata) ToProto() *pb.Metadata {
	return &pb.Metadata{
		Key:   pb.Metadata_OTHER,
		Value: []byte(m.Data),
	}
}

// MarshalText implements the encoding.TextMarshaler interface
func (m OtherMetadata) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("Other %s", m.Data)), nil
}

// MetadataFromProto converts a protocol buffer metadata to a Metadata
func MetadataFromProto(proto *pb.Metadata) Metadata {
	switch proto.Key {
	case pb.Metadata_APP_EUI:
		appEUI := new(types.AppEUI)
		appEUI.UnmarshalBinary(proto.Value)
		return AppEUIMetadata{*appEUI}
	case pb.Metadata_APP_ID:
		return AppIDMetadata{string(proto.Value)}
	case pb.Metadata_PREFIX:
		prefix := types.DevAddrPrefix{
			Length: 32,
		}
		if len(proto.Value) == 5 {
			prefix.Length = int(proto.Value[0])
			prefix.DevAddr[0] = proto.Value[1]
			prefix.DevAddr[1] = proto.Value[2]
			prefix.DevAddr[2] = proto.Value[3]
			prefix.DevAddr[3] = proto.Value[4]
		}
		return PrefixMetadata{prefix}
	case pb.Metadata_OTHER:
		return OtherMetadata{string(proto.Value)}
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
	case "Other":
		return OtherMetadata{value}
	}
	return nil
}

// Announcement of a network component
type Announcement struct {
	old            *Announcement
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
	Metadata       []Metadata

	CreatedAt time.Time `redis:"created_at"`
	UpdatedAt time.Time `redis:"updated_at"`
}

// StartUpdate stores the state of the announcement
func (a *Announcement) StartUpdate() {
	old := *a
	a.old = &old
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
		Metadata:       metadata,
	}
}
