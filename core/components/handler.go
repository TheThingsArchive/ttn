// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/apex/log"
)

var ErrNotImplemented = fmt.Errorf("Ilegal call on non implemented method")

type Handler struct {
	ctx log.Interface
	db  handlerStorage
}

type handlerStorage interface {
	store(data interface{}) error
}

func NewHandler(db handlerStorage, ctx log.Interface) (*Handler, error) {
	return &Handler{
		ctx: ctx,
		db:  db,
	}, nil
}

func (h *Handler) Register(reg core.Registration, an core.AckNacker) error {
	return nil
}

func (h *Handler) HandleUp(p core.Packet, an core.AckNacker, upAdapter core.Adapter) error {
	return nil
}

func (h *Handler) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return ErrNotImplemented
}
