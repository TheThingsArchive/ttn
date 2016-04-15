package models

import (
	"fmt"
	"encoding/json"
	"encoding/hex"
)

type accessKey string
type EUI       []byte

// read EUI hex string into []byte
func ReadEUI(seui string) ([]byte, error) {
	return hex.DecodeString(seui)
}

func WriteEUI(beui []byte) string {
	return hex.EncodeToString(beui)
}

// read JSON eui field into EUI
func (eui *EUI) UnmarshalJSON(data []byte) error {
	bta, err := ReadEUI(string(data[1:17]))
	if err != nil {
		return err
	}
	*eui = EUI(bta)
	return nil
}

// Marshal a EUI to a JSON string
func (eui EUI) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%X\"", eui)), nil
}


type User struct {
	Email string        `json:"email"`
	Apps  []Application `json:"applications"`
}
type user_ User
func (user User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
			user_
			Type string `json:"@type"`
			Self string `json:"@self"`
		}{
			user_:  user_(user),
			Type: "user",
			Self: "/api/v1/me",
		})
}

type SKey []byte
func ReadSKey(sskey string) ([]byte, error) {
	return hex.DecodeString(sskey)
}

func WriteSKey(bskey SKey) string {
	return hex.EncodeToString(bskey)
}

// Marshal a EUI to a JSON string
func (skey *SKey) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%X\"", skey)), nil
}

func (skey *SKey) UnmarshalJSON(data []byte) error {
	read, err := ReadSKey(string(data[1:17]))
	if err != nil {
		return err
	} else {
		*skey = read
		return nil
	}
}
