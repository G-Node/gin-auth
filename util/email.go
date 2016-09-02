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
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"text/template"
	"time"

	"github.com/G-Node/gin-auth/conf"
)

// EmailDispatcher defines an interface for e-mail dispatch.
type EmailDispatcher interface {
	Send(recipient []string, message []byte) error
}

type emailDispatcher struct {
	conf *conf.SmtpCredentials
	send func(string, smtp.Auth, string, []string, []byte) error
}

// Send sets up authentication for e-mail dispatch via smtp and invokes the objects send function.
func (e *emailDispatcher) Send(recipient []string, content []byte) error {
	addr := e.conf.Host + ":" + strconv.Itoa(e.conf.Port)
	auth := smtp.PlainAuth("", e.conf.Username, e.conf.Password, e.conf.Host)
	if e.conf.Mode != "print" && e.conf.Mode != "skip" {
		netCon, err := net.DialTimeout("tcp", addr, time.Second*10)
		if err != nil {
			return err
		}
		if err = netCon.Close(); err != nil {
			return err
		}
	}
	return e.send(addr, auth, e.conf.From, recipient, content)
}

// NewEmailDispatcher returns an instance of emailDispatcher.
// Dependent on the value of config.smtp.Mode the send method will
// print the e-mail content to the commandline (value "print"), do nothing (value "skip")
// or by default send an e-mail via smtp.SendMail.
func NewEmailDispatcher() EmailDispatcher {
	config := conf.GetSmtpCredentials()
	send := smtp.SendMail
	if config.Mode == "print" {
		send = func(addr string, auth smtp.Auth, from string, recipient []string, cont []byte) error {
			fmt.Printf("E-Mail content:\n---\n%s---\n", string(cont))
			return nil
		}
	} else if config.Mode == "skip" {
		send = func(addr string, auth smtp.Auth, from string, recipient []string, cont []byte) error {
			return nil
		}
	}
	return &emailDispatcher{config, send}
}

// MakeEmailTemplate parses a given template into the main email layout template,
// applies the parsed template to the specified content object and returns
// the result as a bytes.Buffer.
func MakeEmailTemplate(fileName string, content interface{}) *bytes.Buffer {
	var doc bytes.Buffer

	mainFile := conf.GetResourceFile("templates", "emaillayout.txt")
	contentFile := conf.GetResourceFile("templates", fileName)
	tmpl, err := template.ParseFiles(mainFile, contentFile)
	if err != nil {
		panic("Error parsing e-mail template: " + err.Error())
	}

	err = tmpl.Execute(&doc, content)
	if err != nil {
		panic("Error executing e-mail template: " + err.Error())
	}

	return &doc
}
