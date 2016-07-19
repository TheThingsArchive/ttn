package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
)

type HTTPError struct {
	Code    int
	Message string
}

func (e HTTPError) Error() string {
	return e.Message
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

func performRequest(server, accessToken, method, URI string, body, res interface{}) (err error) {
	var req *http.Request

	if body != nil {
		// body is not nil, so serialize it and pass it in the request
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

	//
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return HTTPError{
			Code:    resp.StatusCode,
			Message: resp.Status,
		}
	}

	if res != nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(res); err != nil {
			return err
		}

		if _, err := govalidator.ValidateStruct(res); err != nil {
			return fmt.Errorf("Got an illegal response from server: %s", err)
		}
	}

	return nil
}

// GET does a get request to the account server
func GET(server, accessToken, URI string, res interface{}) error {
	return performRequest(server, accessToken, "GET", URI, nil, res)
}

// DELETE does a delete request to the account server
func DELETE(server, accessToken, URI string) error {
	return performRequest(server, accessToken, "DELETE", URI, nil, nil)
}

// POST creates an HTTP Post request to the specified server, with the body
// encoded as JSON
func POST(server, accessToken, URI string, body, res interface{}) error {
	return performRequest(server, accessToken, "POST", URI, body, res)
}

// POST creates an HTTP Put request to the specified server, with the body
// encoded as JSON
func PUT(server, accessToken, URI string, body, res interface{}) error {
	return performRequest(server, accessToken, "PUT", URI, body, res)
}
