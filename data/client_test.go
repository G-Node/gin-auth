// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"github.com/G-Node/gin-auth/util"
	"github.com/pborman/uuid"
	"testing"
)

const (
	uuidClientGin = "8b14d6bb-cae7-4163-bbd1-f3be46e43e31"
)

func TestListClients(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	clients := ListClients()
	if len(clients) != 1 {
		t.Error("Exactly one client expected in list")
	}
}

func TestGetClient(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

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
	defer util.FailOnPanic(t)
	InitTestDb(t)

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

func TestExistsScope(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	exists := CheckScope(util.NewStringSet("repo-read", "repo-write"))
	if !exists {
		t.Error("Scope does not exist")
	}

	exists = CheckScope(util.NewStringSet("repo-read", "something-wrong"))
	if exists {
		t.Error("Scope should not exist")
	}

	exists = CheckScope(util.NewStringSet("something-wrong"))
	if exists {
		t.Error("Scope should not exist")
	}

	exists = CheckScope(util.NewStringSet())
	if exists {
		t.Error("Scope should not exist")
	}
}

func TestDescribeScope(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	desc, ok := DescribeScope(util.NewStringSet("repo-read", "repo-write"))
	if !ok {
		t.Error("Scope description is not complete")
	}
	if s, ok := desc["repo-read"]; !ok || s == "" {
		t.Error("Description for 'repo-read' is missing")
	}
	if s, ok := desc["repo-write"]; !ok || s == "" {
		t.Error("Description for 'repo-write' is missing")
	}

	_, ok = DescribeScope(util.NewStringSet("repo-read", "something-wrong"))
	if ok {
		t.Error("Scope description should not be complete")
	}

	_, ok = DescribeScope(util.NewStringSet())
	if ok {
		t.Error("Scope description should not be complete")
	}
}

func TestCreateClient(t *testing.T) {
	InitTestDb(t)

	id := uuid.NewRandom().String()
	fresh := Client{
		UUID:             id,
		Name:             "gin-foo",
		Secret:           "secret",
		ScopeProvidedMap: map[string]string{"foo-read": "Read access to foo", "foo-write": "Write access to foo"},
		RedirectURIs:     util.NewStringSet("https://foo.com/redirect")}

	err := fresh.Create()
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
	if !check.ScopeProvided().Contains("foo-read") {
		t.Error("Scope should contain 'foo-read'")
	}
	if !check.ScopeProvided().Contains("foo-write") {
		t.Error("Scope should contain 'foo-write")
	}
	if !check.RedirectURIs.Contains("https://foo.com/redirect") {
		t.Error("Redirect URIs should contain 'https://foo.com/redirect'")
	}
}

func TestDeleteClient(t *testing.T) {
	InitTestDb(t)

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
