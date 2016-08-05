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
)

func TestGetQueuedEmails(t *testing.T) {
	InitTestDb(t)
	emails, err := GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) < 1 {
		t.Error("Expected queued e-mails, but result was empty")
	}
}
