// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import "testing"

func TestData_RemoveExpired(t *testing.T) {
	InitTestDb(t)

	RemoveExpired()

	if len(ListGrantRequests()) != 2 {
		t.Errorf("Number of grant requests (%d) does not match expected number", len(ListGrantRequests()))
	}
}
