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

func TestRandomToken(t *testing.T) {
	token := RandomToken()
	if len(token) != 103 {
		t.Error("Token length is expected to be 103")
	}
}
