// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"fmt"
	"io"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	"github.com/brocaar/lorawan"
)

const (
	TYPE_RPACKET byte = iota
	TYPE_BPACKET
	TYPE_HPACKET
	TYPE_APACKET
	TYPE_JPACKET
	TYPE_CPACKET
)

type RPacket interface {
	Packet
	Metadata() Metadata
	Payload() lorawan.PHYPayload
	DevEUI() lorawan.EUI64
}

// rpacket implements the core.RPacket interface
type rpacket struct {
	*baserpacket
}

// NewRPacket construct a new router packet given a payload and metadata
func NewRPacket(payload lorawan.PHYPayload, metadata Metadata) (RPacket, error) {
	if payload.MACPayload == nil {
		return nil, errors.New(errors.Structural, "MACPAyload should not be empty")
	}

	_, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Packet does not carry a MACPayload")
	}

	return &rpacket{
		&baserpacket{
			payload:  payload,
			metadata: metadata,
		},
	}, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p rpacket) MarshalBinary() ([]byte, error) {
	data, err := p.baserpacket.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	return append([]byte{TYPE_RPACKET}, data...), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *rpacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil packet")
	}

	if len(data) < 1 || data[0] != TYPE_RPACKET {
		return errors.New(errors.Structural, "Not a Router packet")
	}

	var isUp bool
	rw := readwriter.New(data[1:])
	rw.Read(func(data []byte) {
		if data[0] == 1 {
			isUp = true
		}
	})

	var dataMetadata []byte
	rw.Read(func(data []byte) { dataMetadata = data })

	var dataPayload []byte
	rw.Read(func(data []byte) { dataPayload = data })

	if rw.Err() != nil {
		return errors.New(errors.Structural, rw.Err())
	}

	p.metadata = Metadata{}
	if err := p.metadata.UnmarshalJSON(dataMetadata); err != nil {
		return errors.New(errors.Structural, err)
	}

	p.payload = lorawan.NewPHYPayload(isUp)
	if err := p.payload.UnmarshalBinary(dataPayload); err != nil {
		return errors.New(errors.Structural, err)
	}

	return nil
}

// String implements the Stringer interface
func (p rpacket) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.metadata.String())
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.payload)
	return str
}

type BPacket interface {
	Packet
	Commands() []lorawan.MACCommand
	DevEUI() lorawan.EUI64
	FCnt() uint32
	Metadata() Metadata
	Payload() []byte
	ValidateMIC(key lorawan.AES128Key) (bool, error)
}

// bpacket implements the core.BPacket interface
type bpacket struct {
	baserpacket
}

// NewBPacket constructs a new broker packets given a payload and metadata
func NewBPacket(payload lorawan.PHYPayload, metadata Metadata) (BPacket, error) {
	if payload.MACPayload == nil {
		return nil, errors.New(errors.Structural, "MACPAyload should not be empty")
	}

	macPayload, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Packet does not carry a MACPayload")
	}

	if len(macPayload.FRMPayload) != 1 {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected exactly 1")
	}

	if _, ok := macPayload.FRMPayload[0].(*lorawan.DataPayload); !ok {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected only data")
	}

	return &bpacket{
		baserpacket: baserpacket{
			payload:  payload,
			metadata: metadata,
		},
	}, nil
}

// FCnt implements the core.BPacket interface
func (p bpacket) FCnt() uint32 {
	return p.payload.MACPayload.(*lorawan.MACPayload).FHDR.FCnt
}

// Payload implements the core.BPacket interface
func (p bpacket) Payload() []byte {
	macPayload := p.baserpacket.payload.MACPayload.(*lorawan.MACPayload)
	return macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes
}

// ValidateMIC implements the core.BPacket interface
func (p bpacket) ValidateMIC(key lorawan.AES128Key) (bool, error) {
	return p.baserpacket.payload.ValidateMIC(key)
}

// Commands implements the core.BPacket interface
func (p bpacket) Commands() []lorawan.MACCommand {
	return p.baserpacket.payload.MACPayload.(*lorawan.MACPayload).FHDR.FOpts
}

// String implements the fmt.Stringer interface
func (p bpacket) String() string {
	return "TODO"
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p bpacket) MarshalBinary() ([]byte, error) {
	data, err := p.baserpacket.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	return append([]byte{TYPE_BPACKET}, data...), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *bpacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil packet")
	}

	if len(data) < 1 || data[0] != TYPE_BPACKET {
		return errors.New(errors.Structural, "Not a Router packet")
	}

	var isUp bool
	rw := readwriter.New(data[1:])
	rw.Read(func(data []byte) {
		if data[0] == 1 {
			isUp = true
		}
	})

	var dataMetadata []byte
	rw.Read(func(data []byte) { dataMetadata = data })

	var dataPayload []byte
	rw.Read(func(data []byte) { dataPayload = data })

	if rw.Err() != nil {
		return errors.New(errors.Structural, rw.Err())
	}

	p.metadata = Metadata{}
	if err := p.metadata.UnmarshalJSON(dataMetadata); err != nil {
		return errors.New(errors.Structural, err)
	}

	p.payload = lorawan.NewPHYPayload(isUp)
	if err := p.payload.UnmarshalBinary(dataPayload); err != nil {
		return errors.New(errors.Structural, err)
	}

	return nil
}

type HPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevEUI() lorawan.EUI64
	Payload() []byte    // FRMPayload
	Metadata() Metadata // TTL on down, DutyCycle + Rssi on Up
}

// hpacket implements the HPacket interface
type hpacket struct {
	*basehpacket
	metadata Metadata
}

// NewHPacket constructs a new Handler packet
func NewHPacket(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload []byte, metadata Metadata) HPacket {
	if payload == nil {
		payload = make([]byte, 0)
	}
	return &hpacket{
		basehpacket: &basehpacket{
			appEUI:  appEUI,
			devEUI:  devEUI,
			payload: payload,
		},
		metadata: metadata,
	}
}

// Metadata implements the core.Metadata interface
func (p hpacket) Metadata() Metadata {
	return p.metadata
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p hpacket) MarshalBinary() ([]byte, error) {
	dataMetadata, err := p.Metadata().MarshalJSON()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	data, err := p.basehpacket.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(append([]byte{TYPE_HPACKET}, data...))
	rw.Write(dataMetadata)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *hpacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil packet")
	}

	if len(data) < 1 || data[0] != TYPE_HPACKET {
		return errors.New(errors.Structural, "Not a Handler packet")
	}

	rw := readwriter.New(data[1:])
	rw.Read(func(data []byte) { copy(p.basehpacket.appEUI[:], data) })
	rw.Read(func(data []byte) { copy(p.basehpacket.devEUI[:], data) })
	rw.Read(func(data []byte) { p.basehpacket.payload = data })
	var dataMetadata []byte
	rw.Read(func(data []byte) { dataMetadata = data })
	if err := p.metadata.UnmarshalJSON(dataMetadata); err != nil {
		return errors.New(errors.Structural, err)
	}
	return nil
}

// String implements the fmt.Stringer interface
func (p hpacket) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.metadata.String())
	str += fmt.Sprintf("\n\tAppEUI:%+x\n,", p.appEUI)
	str += fmt.Sprintf("\n\tDevEUI:%+x\n,", p.devEUI)
	str += fmt.Sprintf("\n\tPayload:%v\n}", p.Payload)
	return str
}

type APacket interface {
	Packet
	Payload() []byte
	Metadata() []Metadata
}

// apacket implements the core.APacket interface
type apacket struct {
	payload  []byte
	metadata []Metadata
}

// NewAPacket constructs a new application packet
func NewAPacket(payload []byte, metadata []Metadata) (APacket, error) {
	if len(payload) == 0 {
		return nil, errors.New(errors.Structural, "Application packet must hold a payload")
	}

	return &apacket{payload: payload, metadata: metadata}, nil
}

// Payload implements the core.APacket interface
func (p apacket) Payload() []byte {
	return p.payload
}

// Metadata implements the core.Metadata interface
func (p apacket) Metadata() []Metadata {
	return p.metadata
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p apacket) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	for _, m := range p.metadata {
		data, err := m.MarshalJSON()
		if err != nil {
			return nil, errors.New(errors.Structural, err)
		}
		rw.Write(data)
	}
	data, err := rw.Bytes()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw = readwriter.New([]byte{TYPE_APACKET})
	rw.Write(p.payload)
	rw.Write(data)

	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *apacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil apacket")
	}

	if len(data) < 1 || data[0] != TYPE_APACKET {
		return errors.New(errors.Structural, "Not an Application packet")
	}

	var dataMetadata []byte
	rw := readwriter.New(data[1:])
	rw.Read(func(data []byte) { p.payload = data })
	rw.Read(func(data []byte) { dataMetadata = data })
	if rw.Err() != nil {
		return errors.New(errors.Structural, rw.Err())
	}

	p.metadata = make([]Metadata, 0)
	rw = readwriter.New(dataMetadata)
	for {
		var dataMetadata []byte
		rw.Read(func(data []byte) { dataMetadata = data })
		if rw.Err() != nil {
			err, ok := rw.Err().(errors.Failure)
			if ok && err.Fault == io.EOF {
				break
			}
			return errors.New(errors.Structural, rw.Err())
		}
		metadata := new(Metadata)
		if err := metadata.UnmarshalJSON(dataMetadata); err != nil {
			return errors.New(errors.Structural, err)
		}

		p.metadata = append(p.metadata, *metadata)
	}

	return nil
}

// String implements the fmt.Stringer interface
func (p apacket) String() string {
	return "TODO"
}

type JPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevEUI() lorawan.EUI64
	DevNonce() [2]byte
	Metadata() Metadata // Rssi + DutyCycle
}

// joinPacket implements the core.JoinPacket interface
type jpacket struct {
	*basehpacket
	metadata Metadata
}

// NewJoinPacket constructs a new JoinPacket
func NewJPacket(appEUI lorawan.EUI64, devEUI lorawan.EUI64, devNonce [2]byte, metadata Metadata) JPacket {
	return &jpacket{
		basehpacket: &basehpacket{
			appEUI:  appEUI,
			devEUI:  devEUI,
			payload: devNonce[:],
		},
		metadata: metadata,
	}
}

// DevNonce implements the core.JoinPacket interface
func (p jpacket) DevNonce() [2]byte {
	return [2]byte{p.basehpacket.payload[0], p.basehpacket.payload[1]}
}

// Metadata implements the core.JoinPacket interface
func (p jpacket) Metadata() Metadata {
	return p.metadata
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p jpacket) MarshalBinary() ([]byte, error) {
	dataMetadata, err := p.Metadata().MarshalJSON()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	data, err := p.basehpacket.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(append([]byte{TYPE_JPACKET}, data...))
	rw.Write(dataMetadata)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *jpacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil packet")
	}

	if len(data) < 1 || data[0] != TYPE_JPACKET {
		return errors.New(errors.Structural, "Not a JoinRequst packet")
	}

	rw := readwriter.New(data[1:])
	rw.Read(func(data []byte) { copy(p.basehpacket.appEUI[:], data) })
	rw.Read(func(data []byte) { copy(p.basehpacket.devEUI[:], data) })
	rw.Read(func(data []byte) { p.basehpacket.payload = data })
	var dataMetadata []byte
	rw.Read(func(data []byte) { dataMetadata = data })
	if err := p.metadata.UnmarshalJSON(dataMetadata); err != nil {
		return errors.New(errors.Structural, err)
	}
	return nil
}

// String implements the fmt.Stringer interface
func (p jpacket) String() string {
	return "TODO"
}

type CPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevEUI() lorawan.EUI64
	Payload() []byte
	NwkSKey() lorawan.AES128Key
}

// acceptpacket implements the core.AcceptPacket interface
type cpacket struct {
	*basehpacket
	nwkSKey lorawan.AES128Key
}

// NewAcceptPacket constructs a new CPacket
func NewCPacket(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload []byte, nwkSKey lorawan.AES128Key) (CPacket, error) {
	if len(payload) == 0 {
		return nil, errors.New(errors.Structural, "Payload cannot be empty")
	}

	return &cpacket{
		basehpacket: &basehpacket{
			appEUI:  appEUI,
			devEUI:  devEUI,
			payload: payload,
		},
		nwkSKey: nwkSKey,
	}, nil
}

// NwkSKey implements the core.AcceptPacket interface
func (p cpacket) NwkSKey() lorawan.AES128Key {
	return p.nwkSKey
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p cpacket) MarshalBinary() ([]byte, error) {
	data, err := p.basehpacket.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	rw := readwriter.New(append([]byte{TYPE_CPACKET}, data...))
	rw.Write(p.nwkSKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *cpacket) UnmarshalBinary(data []byte) error {
	if p == nil {
		return errors.New(errors.Structural, "Cannot unmarshal nil packet")
	}

	if len(data) < 1 || data[0] != TYPE_CPACKET {
		return errors.New(errors.Structural, "Not a JoinAccept packet")
	}

	rw := readwriter.New(data[1:])
	rw.Read(func(data []byte) { copy(p.basehpacket.appEUI[:], data) })
	rw.Read(func(data []byte) { copy(p.basehpacket.devEUI[:], data) })
	rw.Read(func(data []byte) { p.basehpacket.payload = data })
	rw.Read(func(data []byte) { copy(p.nwkSKey[:], data) })
	return rw.Err()
}

// String implements the fmt.Stringer interface
func (p cpacket) String() string {
	return "TODO"
}

// baserpacket is used to compose other packets
type baserpacket struct {
	payload  lorawan.PHYPayload
	metadata Metadata
}

// Metadata implements the core.RPacket interface
func (p baserpacket) Metadata() Metadata {
	return p.metadata
}

// Payload implements the core.RPacket interface
func (p baserpacket) Payload() lorawan.PHYPayload {
	return p.payload
}

// DevEUI implements the core.RPacket interface
func (p baserpacket) DevEUI() lorawan.EUI64 {
	var devEUI lorawan.EUI64
	copy(devEUI[4:], p.payload.MACPayload.(*lorawan.MACPayload).FHDR.DevAddr[:])
	return devEUI
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p baserpacket) MarshalBinary() ([]byte, error) {
	var mtype byte
	switch p.payload.MHDR.MType {
	case lorawan.JoinRequest:
		fallthrough
	case lorawan.UnconfirmedDataUp:
		fallthrough
	case lorawan.ConfirmedDataUp:
		mtype = 1 // Up
	case lorawan.JoinAccept:
		fallthrough
	case lorawan.UnconfirmedDataDown:
		fallthrough
	case lorawan.ConfirmedDataDown:
		mtype = 2 // Down
	default:
		msg := fmt.Sprintf("Unsupported mtype: %s", p.payload.MHDR.MType.String())
		return nil, errors.New(errors.Implementation, msg)
	}

	dataMetadata, err := p.metadata.MarshalJSON()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	dataPayload, err := p.payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(nil)
	rw.Write([]byte{mtype})
	rw.Write(dataMetadata)
	rw.Write(dataPayload)
	return rw.Bytes()
}

// basehpacket is used to compose other packets
type basehpacket struct {
	appEUI  lorawan.EUI64
	devEUI  lorawan.EUI64
	payload []byte
}

func (p basehpacket) AppEUI() lorawan.EUI64 {
	return p.appEUI
}

func (p basehpacket) DevEUI() lorawan.EUI64 {
	return p.devEUI
}

func (p basehpacket) Payload() []byte {
	return p.payload
}

func (p basehpacket) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(p.appEUI)
	rw.Write(p.devEUI)
	rw.Write(p.payload)
	return rw.Bytes()
}
