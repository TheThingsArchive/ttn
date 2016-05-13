// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package collector

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type collectorManagerServer struct {
	collector *collector
}

func (s *collectorManagerServer) GetApplications(ctx context.Context, req *core.GetApplicationsCollectorReq) (*core.GetApplicationsCollectorRes, error) {
	res := new(core.GetApplicationsCollectorRes)

	apps, err := s.collector.appStorage.List()
	if err != nil {
		return res, err
	}

	res.Applications = make([]*core.CollectorApplication, 0, len(apps))
	for _, eui := range apps {
		res.Applications = append(res.Applications, &core.CollectorApplication{
			AppEUI: eui.Bytes(),
		})
	}

	return res, nil
}

func (s *collectorManagerServer) AddApplication(ctx context.Context, req *core.AddApplicationCollectorReq) (*core.AddApplicationCollectorRes, error) {
	res := new(core.AddApplicationCollectorRes)

	var appEUI types.AppEUI
	if err := appEUI.Unmarshal(req.AppEUI); err != nil {
		return res, err
	}

	if err := s.collector.appStorage.Add(appEUI); err != nil {
		return res, err
	}
	if err := s.collector.appStorage.SetAccessKey(appEUI, req.AppAccessKey); err != nil {
		return res, err
	}

	s.collector.startApp(appEUI)

	return res, nil
}

func (s *collectorManagerServer) RemoveApplication(ctx context.Context, req *core.RemoveApplicationCollectorReq) (*core.RemoveApplicationCollectorRes, error) {
	res := new(core.RemoveApplicationCollectorRes)

	var appEUI types.AppEUI
	if err := appEUI.Unmarshal(req.AppEUI); err != nil {
		return res, err
	}

	s.collector.StopApp(appEUI)

	if err := s.collector.appStorage.Remove(appEUI); err != nil {
		return res, err
	}

	return res, nil
}

func (c *collector) RegisterServer(s *grpc.Server) {
	srv := &collectorManagerServer{c}
	core.RegisterCollectorManagerServer(s, srv)
}
