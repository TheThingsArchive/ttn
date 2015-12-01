package protocol

import (
    "testing"
    "bytes"
)

// ------------------------------------------------------------
// ------------------------- Parse (raw []byte) (error, Packet)
// ------------------------------------------------------------

// Parse() with valid raw data and no payload
func TestParse1(t *testing.T) {
    raw := []byte{0x01, 0x14, 0x14, 0x00}
    err, packet := Parse(raw)

    if err != nil {
        t.Errorf("Failed to parse with error: %#v", err)
    }

    if packet.Version != 0x01 {
        t.Errorf("Invalid parsed version: %x", packet.Version)
    }

    if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
        t.Errorf("Invalid parsed token: %x", packet.Token)
    }

    if packet.Identifier != 0x00 {
        t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
    }

    if packet.Payload != nil {
        t.Errorf("Invalid parsed payload: %x", packet.Payload)
    }
}

// Parse() with valid raw data and payload
func TestParse2(t *testing.T) {
    raw := []byte{0x01, 0x14, 0x14, 0x00, 0x42, 0x14, 0x42, 0x14}
    err, packet := Parse(raw)

    if err != nil {
        t.Errorf("Failed to parse with error: %#v", err)
    }

    if packet.Version != 0x01 {
        t.Errorf("Invalid parsed version: %x", packet.Version)
    }

    if !bytes.Equal([]byte{0x14, 0x14}, packet.Token) {
        t.Errorf("Invalid parsed token: %x", packet.Token)
    }

    if packet.Identifier != 0x00 {
        t.Errorf("Invalid parsed identifier: %x", packet.Identifier)
    }

    if !bytes.Equal([]byte{0x42, 0x14, 0x42, 0x14}, packet.Payload) {
        t.Errorf("Invalid parsed payload: %x", packet.Payload)
    }
}

// Parse() with an invalid version number
func TestParse3(t *testing.T) {
    raw := []byte{0x00, 0x14, 0x14, 0x00, 0x42, 0x14, 0x42, 0x14}
    err, _ := Parse(raw)

    if err == nil {
        t.Errorf("Successfully parsed an incorrect version number")
    }
}

// Parse() with an invalid raw message
func TestParse4(t *testing.T) {
    raw1 := []byte{0x01}
    var raw2 []byte
    err1, _ := Parse(raw1)
    err2, _ := Parse(raw2)

    if err1 == nil {
        t.Errorf("Successfully parsed an raw message")
    }

    if err2 == nil {
        t.Errorf("Successfully parsed a nil raw message")
    }
}

// Parse() with an invalid identifier
func TestParse5(t *testing.T) {
    raw := []byte{0x01, 0x14, 0x14, 0xFF, 0x42, 0x14, 0x42, 0x14}
    err, _ := Parse(raw)

    if err == nil {
        t.Errorf("Successfully parsed an incorrect identifier")
    }
}
