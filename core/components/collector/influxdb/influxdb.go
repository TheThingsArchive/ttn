// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package influxdb

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core/collection"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/influxdata/influxdb/client/v2"
)

type influxDBStorage struct {
	client client.Client
}

// NewDataStorage instantiates a new DataStorage for InfluxDB
func NewDataStorage(addr, username, password string) (collection.DataStorage, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	return &influxDBStorage{c}, nil
}

func (i *influxDBStorage) Save(appEUI types.AppEUI, devEUI types.DevEUI, t time.Time, fields map[string]interface{}) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "packets",
		Precision: "us",
	})
	if err != nil {
		return err
	}

	tags := map[string]string{
		"devEUI": devEUI.String(),
	}
	p, err := client.NewPoint(appEUI.String(), tags, fields, t)
	if err != nil {
		return err
	}

	bp.AddPoint(p)
	return i.client.Write(bp)
}

func (i *influxDBStorage) Close() error {
	return i.client.Close()
}
