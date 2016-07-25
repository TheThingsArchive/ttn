// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package account

import (
	"fmt"
	"time"
)

// ListComponents lists all of the users components
func (a *Account) ListComponents() (components []Component, err error) {
	err = a.get("/api/components", &components)
	return components, err
}

// FindComponent finds a comonent of the specified type with the specified id
func (a *Account) FindComponent(typ, id string) (component []Component, err error) {
	err = a.get(fmt.Sprintf("/api/components/%s/%s", typ, id), &component)
	return component, err
}

// FindBroker finds a broker with the specified id
func (a *Account) FindBroker(id string) (component []Component, err error) {
	return a.FindComponent("broker", id)
}

// FindRouter finds a router with the specified id
func (a *Account) FindRouter(id string) (component []Component, err error) {
	return a.FindComponent("router", id)
}

// FindHandler finds a handler with the specified id
func (a *Account) FindHandler(id string) (component []Component, err error) {
	return a.FindComponent("handler", id)
}

type createComponentReq struct {
	ID string `json:"id" valid:"required"`
}

// CreateComponent creates a component with the specified type and id
func (a *Account) CreateComponent(typ, id string) error {
	body := createComponentReq{
		ID: id,
	}
	return a.post(fmt.Sprintf("/api/components/%s", typ), body, nil)
}

// CreateBroker creates a broker with the specified id
func (a *Account) CreateBroker(id string) error {
	return a.CreateComponent("broker", id)
}

// CreateRouter creates a Router with the specified id
func (a *Account) CreateRouter(id string) error {
	return a.CreateComponent("router", id)
}

// CreateHandler creates a handler with the specified id
func (a *Account) CreateHandler(id string) error {
	return a.CreateComponent("handler", id)
}

type componentTokenRes struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// ComponentToken fetches a token for the component with the given
// type and id
func (a *Account) ComponentToken(typ, id string) (token string, err error) {
	var res componentTokenRes
	err = a.get(fmt.Sprintf("/api/components/%s/%s/token", typ, id), &res)
	return res.Token, err
}

// BrokerToken gets the specified brokers token
func (a *Account) BrokerToken(id string) (token string, err error) {
	return a.ComponentToken("broker", id)
}

// RouterToken gets the specified routers token
func (a *Account) RouterToken(id string) (token string, err error) {
	return a.ComponentToken("router", id)
}

// HandlerToken gets the specified handlers token
func (a *Account) HandlerToken(id string) (token string, err error) {
	return a.ComponentToken("handler", id)
}
