// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"fmt"
	"testing"

	"github.com/G-Node/gin-auth/util"
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

func TestEmail_Create(t *testing.T) {
	InitTestDb(t)

	const recipient = "recipient@nowhere.com"
	const content = "content"

	emails, err := GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	num := len(emails)

	email := &Email{}
	err = email.Create(util.NewStringSet(recipient), []byte(content))
	if err != nil {
		t.Error(err.Error())
	}

	emails, err = GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) != num+1 {
		t.Error("Queued e-mail was not properly created")
	}

	// TODO implement create for all modes

	fmt.Printf("'%d', '%s', '%s', '%s', '%s', asByte: '%s', '%s'\n",
		email.Id, email.Mode.String, email.Sender, email.Recipient.Strings()[0],
		string(email.Content), email.Content, email.CreatedAt.String())
}
