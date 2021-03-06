// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"testing"
)

// FailOnPanic can be used in tests in order to recover from
// a panic and make a test fail.
func FailOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal(r)
	}
}
