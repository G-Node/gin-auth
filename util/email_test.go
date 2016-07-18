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
	"github.com/G-Node/gin-auth/conf"
	"net/smtp"
	"strings"
	"testing"
)

func TestMakePlainEmailTemplate(t *testing.T) {
	const from = "sender@example.com"
	const subject = "This is a test message from your conscience!"
	const message = "Give up your evil ways!"
	recipient := []string{"recipient1@example.com", "recipient2@example.com"}

	fields := &EmailStandardFields{}
	fields.From = from
	fields.To = strings.Join(recipient, ", ")
	fields.Subject = subject
	fields.Body = message

	content := MakePlainEmailTemplate(fields).String()
	if strings.Contains(content, "<no value>") {
		t.Errorf("Part of the template was not properly parsed:\n\n%s", content)
	}

	if !strings.Contains(content, "From: "+from) {
		t.Errorf("Sender line is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "To: "+recipient[0]+", "+recipient[1]) {
		t.Errorf("Recipient line is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "Subject: "+subject) {
		t.Errorf("Subject is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "\n"+message+"\n") {
		t.Errorf("Body is malformed or missing:\n\n%s", content)
	}
}

func TestMakeMultipartEmailTemplate(t *testing.T) {
	const code = "activation_code"
	const sender = "sender@example.com"
	const subject = "This is another test message from your conscience!"
	const boundary = "__boundary__"
	const url = "/points/to/nowhere"
	recipient := []string{"recipient@example.com"}

	basicFields := &EmailStandardFields{}
	basicFields.From = sender
	basicFields.To = strings.Join(recipient, ", ")
	basicFields.Subject = subject

	activateFields := &struct {
		*EmailStandardFields
		Code     string
		BaseUrl  string
		Boundary string
	}{basicFields, code, url, boundary}

	content := MakeMultipartEmailTemplate("emailactivate.html", activateFields).String()
	if strings.Contains(content, "<no value>") {
		t.Errorf("Part of the template was not properly parsed:\n\n%s", content)
	}

	if !strings.Contains(content, "From: "+sender) {
		t.Errorf("Sender line is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "To: "+recipient[0]+"\n") {
		t.Errorf("Recipient line is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "Subject: "+subject) {
		t.Errorf("Subject is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "boundary=\""+boundary+"\"") {
		t.Errorf("Boundary definition is malformed or missing:\n\n%s", content)
	}
	if !strings.Contains(content, "--"+boundary) || !strings.Contains(content, "--"+boundary+"--") {
		t.Errorf("Boundary usage is malformed or missing:\n\n%s", content)
	}
}

func TestEmailDispatcher_Send(t *testing.T) {
	const from = "sender@example.com"
	const subject = "This is a test message from your conscience!"
	const message = "Give up your evil ways!"

	recipient := []string{"recipient1@example.com", "recipient2@example.com"}

	fields := &EmailStandardFields{}
	fields.From = from
	fields.To = strings.Join(recipient, ", ")
	fields.Subject = subject
	fields.Body = message

	content := MakePlainEmailTemplate(fields).Bytes()

	f := func(addr string, auth smtp.Auth, from string, recipient []string, cont []byte) error {
		var err error
		content := string(cont)
		if !strings.Contains(content, "\n"+message+"\n") {
			err = fmt.Errorf("Body is malformed or missing:\n%s", content)
		}
		return err
	}

	config := conf.GetSmtpCredentials()

	disp := &emailDispatcher{conf: config, send: f}
	err := disp.Send(recipient, content)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestNewEmailDispatcher(t *testing.T) {
	const from = "sender@example.com"
	const subject = "This is a test message from your conscience!"
	const message = "Give up your evil ways!"

	recipient := []string{"recipient1@example.com", "recipient2@example.com"}

	fields := &EmailStandardFields{}
	fields.From = from
	fields.To = strings.Join(recipient, ", ")
	fields.Subject = subject
	fields.Body = message

	content := MakePlainEmailTemplate(fields).Bytes()

	mail := NewEmailDispatcher()
	err := mail.Send(recipient, content)
	if err != nil {
		t.Error(err.Error())
	}
}
