// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"testing"
	"time"
)

func TestData_RemoveExpired(t *testing.T) {
	InitTestDb(t)

	var grantLifeTime time.Duration = 1
	numTokens := len(ListAccessTokens())
	numGrantReqs := len(ListGrantRequests())
	numSessions := len(ListSessions())

	RemoveExpired(grantLifeTime)

	if len(ListAccessTokens()) != numTokens-1 {
		t.Errorf("Number of access tokens (%d) does not match expected number", numTokens)
	}
	if len(ListGrantRequests()) != numGrantReqs-1 {
		t.Errorf("Number of grant requests (%d) does not match expected number", numGrantReqs)
	}
	if len(ListSessions()) != numSessions-1 {
		t.Errorf("Number of sessions (%d) does not match expected number", numSessions)
	}
}
