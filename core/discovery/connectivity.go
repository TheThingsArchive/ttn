// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type serviceID struct {
	serviceName string
	id          string
}

type serviceStatus struct {
	ctx    context.Context
	cancel context.CancelFunc
	*grpc.ClientConn

	mu              sync.RWMutex
	lastAnnounce    time.Time
	lastStateChange time.Time
	lastAvailable   time.Time
}

func (s *serviceStatus) Available() (available bool) {
	s.mu.RLock()
	available = !s.lastAvailable.Before(s.lastStateChange)
	s.mu.RUnlock()
	return
}

func (d *discovery) startConnectivityMonitor() {
	go func() {
		for {
			if err := d.updateStatus(); err != nil {
				d.Ctx.WithError(err).Warn("Could not update service status")
			}
			time.Sleep(time.Minute)
		}
	}()
}

func (d *discovery) filterAvailable(announcements []*announcement.Announcement) []*announcement.Announcement {
	d.statusMu.RLock()
	defer d.statusMu.RUnlock()
	if d.serviceStatus == nil {
		return announcements
	}
	filtered := make([]*announcement.Announcement, 0, len(announcements))
	for _, a := range announcements {
		if status, ok := d.serviceStatus[serviceID{a.ServiceName, a.ID}]; ok && status.Available() {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func (d *discovery) updateStatus() error {
	d.statusMu.Lock()
	defer d.statusMu.Unlock()
	if d.serviceStatus == nil {
		d.serviceStatus = make(map[serviceID]*serviceStatus)
	}
	announcements, err := d.services.List(nil)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, a := range announcements {
		id := serviceID{a.ServiceName, a.ID}
		status, ok := d.serviceStatus[id]
		if !ok { // newly announced service
			status = &serviceStatus{lastStateChange: now}
			status.ctx, status.cancel = context.WithCancel(context.Background())
			target := strings.Split(a.NetAddress, ",")[0]
			if target != "" {
				var tlsConfig *tls.Config
				host, _, _ := net.SplitHostPort(target)
				caPool := x509.NewCertPool()
				if caPool.AppendCertsFromPEM([]byte(a.Certificate)) {
					tlsConfig = &tls.Config{ServerName: host, RootCAs: caPool}
				}
				backoff := grpc.DefaultBackoffConfig
				backoff.MaxDelay = 10 * time.Minute
				status.ClientConn, err = grpc.Dial(target,
					grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
					grpc.WithBackoffConfig(backoff),
				)
				if err == nil {
					go func() { // start monitoring
						defer status.ClientConn.Close()
						state := status.ClientConn.GetState()
						for {
							if status.ClientConn.WaitForStateChange(status.ctx, state) { // blocking
								status.mu.Lock()
								state = status.ClientConn.GetState()
								status.lastStateChange = time.Now()
								switch state {
								case connectivity.Idle:
								case connectivity.Connecting:
								case connectivity.Ready:
									d.Ctx.Infof("%s %s is available", id.serviceName, id.id)
									status.lastAvailable = status.lastStateChange
								case connectivity.TransientFailure:
								case connectivity.Shutdown:
								}
								status.mu.Unlock()
							} else { // context canceled
								return
							}
						}
					}()
				}
			}
			d.serviceStatus[serviceID{a.ServiceName, a.ID}] = status
		}
		status.lastAnnounce = now
	}

	// stop monitoring old announcements
	for id, status := range d.serviceStatus {
		if status.lastAnnounce != now {
			status.cancel()
			delete(d.serviceStatus, id)
		}
	}

	return nil
}
