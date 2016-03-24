// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestQueue(t *testing.T) {
	{
		Desc(t, "Put then Contain")
		id := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 99}
		q := newPQueue(1)
		q.Put(id)
		Check(t, true, q.Contains(id), "Contains")
	}

	// ----------

	{
		Desc(t, "Put three on size 2")
		id1 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 99}
		id2 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 17, 99}
		id3 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 18, 99}
		q := newPQueue(2)
		q.Put(id1)
		q.Put(id2)
		q.Put(id3)
		Check(t, false, q.Contains(id1), "Contains")
		Check(t, true, q.Contains(id2), "Contains")
		Check(t, true, q.Contains(id3), "Contains")
	}

	// ----------

	{
		Desc(t, "Put -> Remove -> Contain")
		id := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 99}
		q := newPQueue(2)
		q.Put(id)
		q.Remove(id[:17])
		Check(t, false, q.Contains(id), "Contains")
	}

	// ----------

	{
		Desc(t, "Size 2, Put 2, update first, put 1")
		id1 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 99}
		id2 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 17, 99}
		id3 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 18, 99}
		q := newPQueue(2)
		q.Put(id1)
		q.Put(id2)
		q.Put(id1)
		q.Put(id3)
		Check(t, true, q.Contains(id1), "Contains")
		Check(t, false, q.Contains(id2), "Contains")
		Check(t, true, q.Contains(id3), "Contains")
	}

	// ----------

	{
		Desc(t, "size 2 | Put 1 and update 1 several times, remove, then put 2")
		id1 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 99}
		id2 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 17, 99}
		id3 := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 18, 99}
		q := newPQueue(1)
		q.Put(id1)
		q.Put(id1)
		q.Put(id1)
		q.Put(id1)
		q.Put(id1)
		q.Remove(id1)
		q.Put(id2)
		q.Put(id3)
		Check(t, false, q.Contains(id1), "Contains")
		Check(t, false, q.Contains(id2), "Contains")
		Check(t, true, q.Contains(id3), "Contains")
	}
}
