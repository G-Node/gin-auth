// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"fmt"
	"net/smtp"
	"strings"
	"testing"
)

func TestMakePlainEmailTemplate(t *testing.T) {
	const sender = "sender@example.com"
	const subject = "This is a test message from your conscience!"
	const message = "Give up your evil ways!"
	recipient := []string{"recipient1@example.com", "recipient2@example.com"}

	fields := &EmailStandardFields{}
	fields.From = sender
	fields.To = strings.Join(recipient, ", ")
	fields.Subject = subject
	fields.Body = message

	content := MakePlainEmailTemplate(fields).String()

	if !strings.Contains(content, "From: "+sender) {
		t.Errorf("Sender line is malformed or missing:\n'%s'", content)
	}
	if !strings.Contains(content, "To: "+recipient[0]+", "+recipient[1]) {
		t.Errorf("Recipient line is malformed or missing:\n'%s'", content)
	}
	if !strings.Contains(content, "Subject: "+subject) {
		t.Errorf("Subject is malformed or missing:\n'%s'", content)
	}
	if !strings.Contains(content, "\n"+message+"\n") {
		t.Errorf("Body is malformed or missing:\n\n%s", content)
	}
}

func TestEmailDispatcher_Send(t *testing.T) {
	const identity = ""
	const dispatcher = "dispatcher@some.host.com"
	const pw = "somepw"
	const host = "some.host.com"
	const port = "587"
	const sender = "sender@example.com"
	const subject = "This is a test message from your conscience!"
	const message = "Give up your evil ways!"

	recipient := []string{"recipient1@example.com", "recipient2@example.com"}

	fields := &EmailStandardFields{}
	fields.From = sender
	fields.To = strings.Join(recipient, ", ")
	fields.Subject = subject
	fields.Body = message

	content := MakePlainEmailTemplate(fields).Bytes()

	config := EmailConfig{identity, dispatcher, pw, host, port}

	f := func(addr string, auth smtp.Auth, sender string, recipient []string, cont []byte) error {
		var err error
		content := string(cont)
		if !strings.Contains(content, "\n"+message+"\n") {
			err = fmt.Errorf("Body is malformed or missing:\n%s", content)
		}
		return err
	}

	disp := &emailDispatcher{conf: config, send: f}
	err := disp.Send(recipient, content)
	if err != nil {
		t.Error(err.Error())
	}
}
