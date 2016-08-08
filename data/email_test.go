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
	"strings"
	"testing"

	"github.com/G-Node/gin-auth/conf"
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

func TestEmail_Delete(t *testing.T) {
	InitTestDb(t)

	emails, err := GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) < 1 {
		t.Error("Expected queued e-mails, but result was empty")
	}
	num := len(emails)

	err = emails[0].Delete()
	if err != nil {
		t.Errorf("Error trying to delete e-mail (Id %d, mode '%s'): %s",
			emails[0].Id, emails[0].Mode.String, err.Error())
	}
	emails, err = GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) != num-1 {
		t.Errorf("Number of e-mail entries should be '%d', but was '%d'", num-1, len(emails))
	}
}

func TestEmail_Send(t *testing.T) {
	InitTestDb(t)

	emails, err := GetQueuedEmails()
	if err != nil {
		t.Errorf("Error fetching queued e-mails: '%s'\n", err.Error())
	}
	if len(emails) < 3 {
		t.Error("Expected queued e-mails, but result did not match")
	}

	// test send e-mail, check bad username error
	if emails[0].Mode.Valid {
		t.Error("Expected e-mail mode to be empty")
	}

	username := conf.GetSmtpCredentials().Username
	conf.GetSmtpCredentials().Username = "iDoNotExist"
	err = emails[0].Send()
	if err == nil {
		t.Error("Expected error")
	}
	if !strings.Contains(err.Error(), "Bad username or password") {
		t.Errorf("Expected Bad username error but got: '%s'", err.Error())
	}
	// reset username
	conf.GetSmtpCredentials().Username = username

	// test print option
	if emails[1].Mode.String != "print" {
		t.Error("Expected e-mail mode to be print")
	}
	err = emails[1].Send()
	if err != nil {
		t.Errorf("Unexpected error when printing e-mail: '%s'", err.Error())
	}

	// test skip option
	if emails[2].Mode.String != "skip" {
		t.Error("Expected e-mail mode to be skip")
	}
	err = emails[2].Send()
	if err != nil {
		t.Errorf("Unexpected error when skipping e-mail: '%s'", err.Error())
	}
}
