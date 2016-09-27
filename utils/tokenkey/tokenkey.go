// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package tokenkey

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
)

// K is the data returned by the token key provider
type K struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
}

// Provider represents a provider of the token key
type Provider interface {
	Get(server string, renew bool) (*K, error)
	Update() error
}

type httpProvider struct {
	servers       map[string]string
	cache         map[string][]byte
	cacheLocation string
}

// NewHTTPProvider returns a new Provider that fetches the key from a HTTP
// resource
func NewHTTPProvider(servers map[string]string, cacheLocation string) Provider {
	return &httpProvider{servers, map[string][]byte{}, cacheLocation}
}

func (p *httpProvider) Get(server string, renew bool) (*K, error) {
	var data []byte

	url, ok := p.servers[server]
	if !ok {
		return nil, fmt.Errorf("Auth server %s not registered", server)
	}

	cacheFile := path.Join(p.cacheLocation, fmt.Sprintf("auth-%s.pub", server))

	// Try to read from memory
	cached, ok := p.cache[server]
	if ok {
		data = cached
	}

	if data == nil {
		// Try to read from cache file
		cached, err := ioutil.ReadFile(cacheFile)
		if err == nil {
			p.cache[server] = cached
			data = cached
		}
	}

	// Fetch token if there's a renew or if there's no key cached
	if renew || data == nil {
		fetched, err := p.fetch(fmt.Sprintf("%s/key", url))
		if err == nil {
			data = fetched
			// Don't care about errors here. It's better to retrieve keys all the time
			// because they can't be cached than not to be able to verify a token
			ioutil.WriteFile(cacheFile, data, 0644)
			p.cache[server] = data
		} else if data == nil {
			return nil, err // We don't have a key here
		}
	}

	var key K
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, err
	}

	return &key, nil
}

func (p *httpProvider) Update() error {
	for server := range p.servers {
		_, err := p.Get(server, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *httpProvider) fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
