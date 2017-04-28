// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package component

import "sort"

const defaultPriority = 0

type initFunc func(c *Component) error
type initFuncs map[int][]initFunc

func (i initFuncs) add(f initFunc) {
	i.addAt(defaultPriority, f)
}

func (i initFuncs) addAt(priority int, f initFunc) {
	if _, ok := i[priority]; !ok {
		i[priority] = make([]initFunc, 0)
	}
	i[priority] = append(i[priority], f)
}

func (i initFuncs) run(c *Component) error {
	var keys []int
	for key := range i {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		for _, init := range i[key] {
			if err := init(c); err != nil {
				return err
			}
		}
	}
	return nil
}

var _initFuncs initFuncs = make(map[int][]initFunc)

// OnInitialize registers a function that is called when any Component is initialized
func OnInitialize(fun initFunc) {
	_initFuncs.add(fun)
}

func (c *Component) initialize() error {
	return _initFuncs.run(c)
}
