// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package conf

import (
	"testing"
)

func TestGetServerConfig(t *testing.T) {
	config := GetServerConfig()
	if config.Host != "localhost" {
		t.Error("Host expected to be 'localhost'")
	}
	if config.Port != 8081 {
		t.Error("Port expected to be '8081'")
	}
	if config.BaseURL != "http://localhost:8081" {
		t.Error("BaseURL expected to be 'http://localhost:8081'")
	}
}

func TestGetDbConfig(t *testing.T) {
	config := GetDbConfig()
	if config.Driver != "postgres" {
		t.Error("Driver expected to be 'postgres'")
	}
}

func TestGetSmtpCredentials(t *testing.T) {
	creds := GetSmtpCredentials()
	if creds.From != "no-reply@g-node.org" {
		t.Errorf("From expected to be no-reply@g-node.org but was '%s'", creds.From)
	}
}
