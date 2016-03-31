package tokenkey

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// K is the data returned by the token key provider
type K struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
}

// Provider represents a provider of the token key
type Provider interface {
	fmt.Stringer
	Get(renew bool) (*K, error)
}

type httpProvider struct {
	url       string
	cacheFile string
}

// NewHTTPProvider returns a new Provider that fetches the key from a HTTP
// resource
func NewHTTPProvider(url, cacheFile string) Provider {
	return &httpProvider{url, cacheFile}
}

func (p *httpProvider) String() string {
	return p.url
}

func (p *httpProvider) Get(renew bool) (*K, error) {
	var data []byte

	// Try to read from cache
	cached, err := ioutil.ReadFile(p.cacheFile)
	if err == nil {
		data = cached
	}

	// Fetch token if there's a renew or if there's no key cached
	if renew || data == nil {
		fetched, err := p.fetch()
		if err == nil {
			data = fetched
			// Don't care about errors here. It's better to retrieve keys all the time
			// because they can't be cached than not to be able to verify a token
			ioutil.WriteFile(p.cacheFile, data, 0644)
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

func (p *httpProvider) fetch() ([]byte, error) {
	resp, err := http.Get(p.url)
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
