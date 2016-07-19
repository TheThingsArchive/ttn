package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// NewRequest creates a new http.Request that has authorization set up
func newRequest(server string, accessToken string, method string, URI string, body io.Reader) (*http.Request, error) {
	URL := fmt.Sprint("%s%s", server, URI)
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// GET does a get request to the account server
func GET(server, accessToken, URI string) (*http.Response, error) {
	req, err := c.NewRequest(server, accessToken, "GET", URI, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DELETE does a delete request to the account server
func DELETE(server, accessToken, URI string) (*http.Response, error) {
	req, err := c.NewRequest(server, accessToken, "DELETE", URI, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// POST creates an HTTP Post request to the specified server, with the body
// encoded as JSON
func POST(server, accessToken, URI string, body interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := c.NewRequest(server, accessToken, "POST", URI, buf)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// POST creates an HTTP Put request to the specified server, with the body
// encoded as JSON
func PUT(server, accessToken, URI string, body interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(body)
	if err != nil {
		return nil, err
	}
	req, err := c.NewRequest(server, accessToken, "PUT", URI, buf)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
