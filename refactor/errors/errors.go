// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package errors

const (
	// Unsuccessful operation due to an unexpected parameter or under-relying structure.
	// Another call won't change a thing, the inputs are wrong anyway.
	ErrInvalidStructure = "Invalid Structure"

	// Attempt to access an unimplemented method or an unsupported operation. Fatal.
	ErrNotSupported = "Unsupported Operation"

	// The operation went well though the result is unexpected and wrong.
	ErrWrongBehavior = "Unexpected Behavior"

	// Something happend during the processing. Another attempt might success.
	ErrFailedOperation = "Unsuccessful Operation"
)
