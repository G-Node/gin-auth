// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"strings"
	"testing"
)

func TestMakePlainEmailTemplate(t *testing.T) {
	const sender = "sender@example.com"
	const subject = "This is a test message from your conscience!"
	const message = "Give up your evil ways!"
	recipient := []string{"recipient1@example.com", "recipient2@example.com"}

	body := MakePlainEmailTemplate(sender, recipient, subject, message).String()

	if !strings.Contains(body, "From: "+sender) {
		t.Errorf("Sender line is malformed or missing:\n'%s'", body)
	}
	if !strings.Contains(body, "To: "+recipient[0]+", "+recipient[1]) {
		t.Errorf("Recipient line is malformed or missing:\n'%s'", body)
	}
	if !strings.Contains(body, "Subject: "+subject) {
		t.Errorf("Subject is malformed or missing:\n'%s'", body)
	}
	if !strings.Contains(body, "\n"+message+"\n") {
		t.Errorf("Body is malformed or missing:\n'%s'", body)
	}
}
