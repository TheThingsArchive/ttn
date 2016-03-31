package tokenkey

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// T is the data returned by the token key provider
type T struct {
	Algorithm string `json:"algorithm"`
	Key       string `json:"key"`
}

// Provider represents a provider of the token key
type Provider interface {
	fmt.Stringer
	Get() (*T, error)
	Refresh() (*T, error)
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

func (p *httpProvider) Get() (*T, error) {
	var data []byte

	// Try to read the data from cache
	d, err := ioutil.ReadFile(p.cacheFile)
	if err == nil {
		data = d
	}

	// If the file doesn't exist or if there's a read error, get it from the
	// server
	if data == nil {
		resp, err := http.Get(p.url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(resp.Status)
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	var token T
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	// Don't care about errors here. It's better to retrieve keys all the time
	// because they can't be cached than not to be able to verify a token
	ioutil.WriteFile(p.cacheFile, data, 0644)

	return &token, nil
}

func (p *httpProvider) Refresh() (*T, error) {
	// Just delete the cached file...
	if err := os.Remove(p.cacheFile); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// ...so that Get always gets a new file
	return p.Get()
}
