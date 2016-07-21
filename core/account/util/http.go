// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// MaxRedirects specifies the maximum number of redirects an HTTP
	// request should be able to make
	MaxRedirects = 5
)

// HTTPError represents an error coming over HTTP,
// it is not an error with executing the request itself, it is
// an error the server is flaggin to the client.
type HTTPError struct {
	Code    int
	Message string
}

func (e HTTPError) Error() string {
	return e.Message
}

// checkRedirect implements this clients redirection policy
func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > MaxRedirects {
		return errors.New("Maximum number of redirects reached")
	}

	// use the same headers as before
	req.Header.Set("Authorization", via[len(via)-1].Header.Get("Authorization"))
	req.Header.Set("Accept", "application/json")
	return nil
}

// NewRequest creates a new http.Request that has authorization set up
func newRequest(server string, accessToken string, method string, URI string, body io.Reader) (*http.Request, error) {
	URL := fmt.Sprintf("%s%s", server, URI)
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func performRequest(server, accessToken, method, URI string, body, res interface{}, redirects int) (err error) {
	var req *http.Request

	if body != nil {
		// body is not nil, so serialize it and pass it in the request
		if err = Validate(body); err != nil {
			return fmt.Errorf("Got an illegal request body: %s", err)
		}

		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		err = encoder.Encode(body)
		if err != nil {
			return err
		}
		req, err = newRequest(server, accessToken, method, URI, buf)
	} else {
		// body is nil so create a nil request
		req, err = newRequest(server, accessToken, method, URI, nil)
	}

	if err != nil {
		return err
	}

	client := &http.Client{
		CheckRedirect: checkRedirect,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return HTTPError{
			Code:    resp.StatusCode,
			Message: resp.Status,
		}
	}

	if resp.StatusCode == 307 {
		if redirects > 0 {
			location := resp.Header.Get("Location")
			return performRequest(server, accessToken, method, location, body, res, redirects-1)
		}
		return fmt.Errorf("Reached maximum number of redirects")
	}

	if resp.StatusCode >= 300 {
		// 307 is handled, 301, 302, 304 cannot be
		return fmt.Errorf("Unexpected %v redirection to %s", resp.StatusCode, resp.Header.Get("Location"))
	}

	if res != nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(res); err != nil {
			return err
		}

		if err := Validate(res); err != nil {
			return fmt.Errorf("Got an illegal response from server: %s", err)
		}
	}

	return nil
}

// GET does a get request to the account server,  decoding the result into the object pointed to byres
func GET(server, accessToken, URI string, res interface{}) error {
	return performRequest(server, accessToken, "GET", URI, nil, res, MaxRedirects)
}

// DELETE does a delete request to the account server
func DELETE(server, accessToken, URI string) error {
	return performRequest(server, accessToken, "DELETE", URI, nil, nil, MaxRedirects)
}

// POST creates an HTTP Post request to the specified server, with the body
// encoded as JSON, decoding the result into the object pointed to byres
func POST(server, accessToken, URI string, body, res interface{}) error {
	return performRequest(server, accessToken, "POST", URI, body, res, MaxRedirects)
}

// PUT creates an HTTP Put request to the specified server, with the body
// encoded as JSON, decoding the result into the object pointed to byres
func PUT(server, accessToken, URI string, body, res interface{}) error {
	return performRequest(server, accessToken, "PUT", URI, body, res, MaxRedirects)
}

// PATCH creates an HTTP Patch request to the specified server, with the body
// encoded as JSON, decoding the result into the object pointed to byres
func PATCH(server, accessToken, URI string, body, res interface{}) error {
	return performRequest(server, accessToken, "POST", URI, body, res, MaxRedirects)
}
