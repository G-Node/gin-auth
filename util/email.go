// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"bytes"
	"strings"
	"text/template"
)

// MakePlainEmailTemplate returns a bytes.Buffer containing a standard e-mail
func MakePlainEmailTemplate(from string, to []string, subj string, messageBody string) *bytes.Buffer {
	const emailTemplate = `From: {{ .From }}
To: {{ .To }}
Subject: {{ .Subject }}

{{ .Body }}
`
	var doc bytes.Buffer

	content := &struct {
		From    string
		To      string
		Subject string
		Body    string
	}{
		from,
		strings.Join(to, ", "),
		subj,
		messageBody,
	}
	t := template.New("emailTemplate")
	t, err := t.Parse(emailTemplate)
	if err != nil {
		panic("Error parsing e-mail template")
	}
	err = t.Execute(&doc, content)
	if err != nil {
		panic("Error executing e-mail template")
	}
	return &doc
}
