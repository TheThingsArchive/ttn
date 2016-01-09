// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broadcast

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	httpadapter "github.com/thethingsnetwork/core/adapters/http"
	"github.com/thethingsnetwork/core/utils/log"
)

type Adapter struct {
	*httpadapter.Adapter
	recipients []core.Recipient
}

var ErrBadOptions = fmt.Errorf("Bad options provided")

func NewAdapter(recipients []core.Recipient, loggers ...log.Logger) (*Adapter, error) {
	if len(recipients) == 0 {
		return nil, ErrBadOptions
	}

	adapter, err := httpadapter.NewAdapter(loggers...)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		Adapter:    adapter,
		recipients: recipients,
	}, nil
}

func (a *Adapter) broadcast(p core.Packet) error {
	return nil
}
