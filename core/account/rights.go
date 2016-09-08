// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TheThingsNetwork/ttn/core/account/util"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func AccessKeyRights(server string, appID string, key string) ([]types.Right, error) {

	url := fmt.Sprintf("%s/applications/%s/rights", server, appID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Key %s", key))

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.StatusCode != 200 {
		var herr util.HTTPError
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&herr); err != nil {
			// could not decode body as error, just return http error
			return nil, util.HTTPError{
				Code:    resp.StatusCode,
				Message: resp.Status[4:],
			}
		}

		// fill in blank code
		if herr.Code == 0 {
			herr.Code = resp.StatusCode
		}

		// fill in blank message
		if herr.Message == "" {
			herr.Message = resp.Status[4:]
		}

		return nil, herr
	}
	defer resp.Body.Close()

	var rights []types.Right
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&rights); err != nil {
		return nil, err
	}

	return rights, nil
}
