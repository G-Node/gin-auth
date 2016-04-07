// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"github.com/pborman/uuid"
	"testing"
)

const (
	uuidClientGin = "8b14d6bb-cae7-4163-bbd1-f3be46e43e31"
)

func TestListClients(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	clients := ListClients()
	if len(clients) != 1 {
		t.Error("Exactly one client expected in list")
	}
}

func TestGetClient(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	client, ok := GetClient(uuidClientGin)
	if !ok {
		t.Error("Client does not exist")
	}
	if client.Name != "gin" {
		t.Error("Client name was expected to be 'gin'")
	}

	_, ok = GetClient("doesNotExist")
	if ok {
		t.Error("Client should not exist")
	}
}

func TestGetClientByName(t *testing.T) {
	defer failOnPanic(t)
	initTestDb(t)

	client, ok := GetClientByName("gin")
	if !ok {
		t.Error("Client does not exist")
	}
	if client.UUID != uuidClientGin {
		t.Errorf("Client UUID was expected to be '%s'", uuidClientGin)
	}

	_, ok = GetClientByName("doesNotExist")
	if ok {
		t.Error("Client should not exist")
	}
}

func TestCreateClient(t *testing.T) {
	initTestDb(t)

	id := uuid.NewRandom().String()
	new := Client{
		UUID:          id,
		Name:          "gin-foo",
		Secret:        "secret",
		ScopeProvided: SqlStringSlice{"foo-read", "foo-write"},
		RedirectURIs:  SqlStringSlice{"https://foo.com/redirect"}}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, ok := GetClient(id)
	if !ok {
		t.Error("Client does not exist")
	}
	if check.Name != "gin-foo" {
		t.Error("Name was expected to bo 'gin-foo'")
	}
	if check.ScopeProvided[0] != "foo-read" {
		t.Error("First scope was expected to be 'foo-read'")
	}
	if check.ScopeProvided[1] != "foo-write" {
		t.Error("Second scope was expected to be 'foo-write")
	}
	if check.RedirectURIs[0] != "https://foo.com/redirect" {
		t.Error("First redirect was expected to be 'https://foo.com/redirect'")
	}
}

func TestDeleteClient(t *testing.T) {
	initTestDb(t)

	client, ok := GetClient(uuidClientGin)
	if !ok {
		t.Error("Client does not exist")
	}

	err := client.Delete()
	if err != nil {
		t.Error(err)
	}

	_, ok = GetClient(uuidClientGin)
	if ok {
		t.Error("Client should not exist")
	}
}
