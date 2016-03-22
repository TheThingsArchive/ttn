package core

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// MarshalMsg implements msgp.Marshaler
func (z ABPSubAppReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "dev_addr"
	o = append(o, 0x83, 0xa8, 0x64, 0x65, 0x76, 0x5f, 0x61, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.DevAddr)
	// string "nwks_key"
	o = append(o, 0xa8, 0x6e, 0x77, 0x6b, 0x73, 0x5f, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.NwkSKey)
	// string "apps_key"
	o = append(o, 0xa8, 0x61, 0x70, 0x70, 0x73, 0x5f, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.AppSKey)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ABPSubAppReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "dev_addr":
			z.DevAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "nwks_key":
			z.NwkSKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "apps_key":
			z.AppSKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z ABPSubAppReq) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.DevAddr) + 9 + msgp.StringPrefixSize + len(z.NwkSKey) + 9 + msgp.StringPrefixSize + len(z.AppSKey)
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AppMetadata) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 9
	// string "frequency"
	o = append(o, 0x89, 0xa9, 0x66, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79)
	o = msgp.AppendFloat32(o, z.Frequency)
	// string "data_rate"
	o = append(o, 0xa9, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x72, 0x61, 0x74, 0x65)
	o = msgp.AppendString(o, z.DataRate)
	// string "coding_rate"
	o = append(o, 0xab, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x5f, 0x72, 0x61, 0x74, 0x65)
	o = msgp.AppendString(o, z.CodingRate)
	// string "timestamp"
	o = append(o, 0xa9, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o = msgp.AppendUint32(o, z.Timestamp)
	// string "rssi"
	o = append(o, 0xa4, 0x72, 0x73, 0x73, 0x69)
	o = msgp.AppendInt32(o, z.Rssi)
	// string "lsnr"
	o = append(o, 0xa4, 0x6c, 0x73, 0x6e, 0x72)
	o = msgp.AppendFloat32(o, z.Lsnr)
	// string "altitude"
	o = append(o, 0xa8, 0x61, 0x6c, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65)
	o = msgp.AppendInt32(o, z.Altitude)
	// string "longitude"
	o = append(o, 0xa9, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65)
	o = msgp.AppendFloat32(o, z.Longitude)
	// string "latitude"
	o = append(o, 0xa8, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65)
	o = msgp.AppendFloat32(o, z.Latitude)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AppMetadata) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "frequency":
			z.Frequency, bts, err = msgp.ReadFloat32Bytes(bts)
			if err != nil {
				return
			}
		case "data_rate":
			z.DataRate, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "coding_rate":
			z.CodingRate, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "timestamp":
			z.Timestamp, bts, err = msgp.ReadUint32Bytes(bts)
			if err != nil {
				return
			}
		case "rssi":
			z.Rssi, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				return
			}
		case "lsnr":
			z.Lsnr, bts, err = msgp.ReadFloat32Bytes(bts)
			if err != nil {
				return
			}
		case "altitude":
			z.Altitude, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				return
			}
		case "longitude":
			z.Longitude, bts, err = msgp.ReadFloat32Bytes(bts)
			if err != nil {
				return
			}
		case "latitude":
			z.Latitude, bts, err = msgp.ReadFloat32Bytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *AppMetadata) Msgsize() (s int) {
	s = 1 + 10 + msgp.Float32Size + 10 + msgp.StringPrefixSize + len(z.DataRate) + 12 + msgp.StringPrefixSize + len(z.CodingRate) + 10 + msgp.Uint32Size + 5 + msgp.Int32Size + 5 + msgp.Float32Size + 9 + msgp.Int32Size + 10 + msgp.Float32Size + 9 + msgp.Float32Size
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *DataDownAppReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "payload"
	o = append(o, 0x82, 0xa7, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	o = msgp.AppendBytes(o, z.Payload)
	// string "ttl"
	o = append(o, 0xa3, 0x74, 0x74, 0x6c)
	o = msgp.AppendString(o, z.TTL)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DataDownAppReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "payload":
			z.Payload, bts, err = msgp.ReadBytesBytes(bts, z.Payload)
			if err != nil {
				return
			}
		case "ttl":
			z.TTL, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *DataDownAppReq) Msgsize() (s int) {
	s = 1 + 8 + msgp.BytesPrefixSize + len(z.Payload) + 4 + msgp.StringPrefixSize + len(z.TTL)
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *DataUpAppReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "payload"
	o = append(o, 0x82, 0xa7, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64)
	o = msgp.AppendBytes(o, z.Payload)
	// string "metadata"
	o = append(o, 0xa8, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Metadata)))
	for xvk := range z.Metadata {
		o, err = z.Metadata[xvk].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DataUpAppReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "payload":
			z.Payload, bts, err = msgp.ReadBytesBytes(bts, z.Payload)
			if err != nil {
				return
			}
		case "metadata":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Metadata) >= int(xsz) {
				z.Metadata = z.Metadata[:xsz]
			} else {
				z.Metadata = make([]AppMetadata, xsz)
			}
			for xvk := range z.Metadata {
				bts, err = z.Metadata[xvk].UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *DataUpAppReq) Msgsize() (s int) {
	s = 1 + 8 + msgp.BytesPrefixSize + len(z.Payload) + 9 + msgp.ArrayHeaderSize
	for xvk := range z.Metadata {
		s += z.Metadata[xvk].Msgsize()
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *OTAAAppReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "metadata"
	o = append(o, 0x81, 0xa8, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Metadata)))
	for bzg := range z.Metadata {
		o, err = z.Metadata[bzg].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *OTAAAppReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "metadata":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Metadata) >= int(xsz) {
				z.Metadata = z.Metadata[:xsz]
			} else {
				z.Metadata = make([]AppMetadata, xsz)
			}
			for bzg := range z.Metadata {
				bts, err = z.Metadata[bzg].UnmarshalMsg(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *OTAAAppReq) Msgsize() (s int) {
	s = 1 + 9 + msgp.ArrayHeaderSize
	for bzg := range z.Metadata {
		s += z.Metadata[bzg].Msgsize()
	}
	return
}
