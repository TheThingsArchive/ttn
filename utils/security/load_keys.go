package security

import (
	"io/ioutil"
	"path/filepath"
)

func LoadKeys(location string) (pubKey, privKey, cert []byte, err error) {
	pubKey, err = ioutil.ReadFile(filepath.Clean(location + "/server.pub"))
	if err != nil {
		return
	}
	privKey, err = ioutil.ReadFile(filepath.Clean(location + "/server.key"))
	if err != nil {
		return
	}
	cert, err = ioutil.ReadFile(filepath.Clean(location + "/server.cert"))
	if err != nil {
		return
	}
	return
}
