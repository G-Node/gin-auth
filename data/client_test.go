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

// TODO remove function
func TestListClients(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	clients := ListClients()
	if len(clients) != 1 {
		t.Error("Exactly one client expected in list")
	}
}

func TestListClientUUIDs(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	clientList := listClientUUIDs()
	if len(clientList) != 1 && !clientList.Contains(uuidClientGin) {
		t.Error("listClientUUIDs returned incomplete list.")
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

func TestClientCreate(t *testing.T) {
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

func TestClientApprovalForAccountAndApprove(t *testing.T) {
	InitTestDb(t)

	client, ok := GetClient(uuidClientGin)
	if !ok {
		t.Error("Client does not exist")
	}

	approval, ok := client.ApprovalForAccount(uuidAlice)
	if !ok {
		t.Error("Approval does not exist")
	}
	if approval.AccountUUID != uuidAlice {
		t.Error("Wrong account uuid")
	}
	if approval.ClientUUID != client.UUID {
		t.Error("Wrong client uuid")
	}

	approval, ok = client.ApprovalForAccount(uuidBob)
	if ok {
		t.Error("Approval should not exist")
	}

	client.Approve(uuidBob, util.NewStringSet("repo-read"))
	approval, ok = client.ApprovalForAccount(uuidBob)
	if !ok {
		t.Error("Approval does not exist")
	}
	if !approval.Scope.Contains("repo-read") {
		t.Error("Approval scope should contain 'repo-read'")
	}
	if approval.Scope.Contains("repo-write") {
		t.Error("Approval scope should not contain 'repo-write'")
	}

	client.Approve(uuidBob, util.NewStringSet("repo-read", "repo-write"))
	approval, ok = client.ApprovalForAccount(uuidBob)
	if !ok {
		t.Error("Approval does not exist")
	}
	if !approval.Scope.Contains("repo-read") {
		t.Error("Approval scope should contain 'repo-read'")
	}
	if !approval.Scope.Contains("repo-write") {
		t.Error("Approval scope should contain 'repo-write'")
	}
}

func TestClientCreateGrantRequest(t *testing.T) {
	InitTestDb(t)

	client, ok := GetClient(uuidClientGin)
	if !ok {
		t.Error("Client does not exist")
	}

	state := util.RandomToken()

	// wrong response type
	request, err := client.CreateGrantRequest("foo", client.RedirectURIs.Strings()[0], state, util.NewStringSet("repo-read"))
	if err == nil {
		t.Error("Error expected")
	}

	// wrong redirect
	request, err = client.CreateGrantRequest("foo", "https://doesnotexist.com/callback", state, util.NewStringSet("repo-read"))
	if err == nil {
		t.Error("Error expected")
	}

	// wrong scope
	request, err = client.CreateGrantRequest("foo", client.RedirectURIs.Strings()[0], state, util.NewStringSet("foo-read"))
	if err == nil {
		t.Error("Error expected")
	}

	// all OK
	request, err = client.CreateGrantRequest("code", client.RedirectURIs.Strings()[0], state, util.NewStringSet("repo-read"))
	if err != nil {
		t.Error(err)
	}
	if request.ClientUUID != client.UUID {
		t.Error("Client UUID does not match")
	}
	if !request.ScopeRequested.Contains("repo-read") {
		t.Error("The requested scope should contain 'repo-read'")
	}
	if request.State != state {
		t.Error("State does not match")
	}
}

func TestClientScopeProvided(t *testing.T) {
	InitTestDb(t)

	client := &Client{ScopeProvidedMap: map[string]string{"foo": "Foo", "bar": "Bar"}}
	scope := client.ScopeProvided()

	if scope.Len() != 2 {
		t.Errorf("Scope should have 2 elements but has %d", scope.Len())
	}
	if !scope.Contains("foo") {
		t.Error("Scope should contain 'foo'")
	}
	if !scope.Contains("bar") {
		t.Error("Scope should contain 'bar'")
	}
}

func TestClientDelete(t *testing.T) {
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

// Tests that InitClients panics correctly, if the clients
// file does not exist.
func TestInitClientsMissingFile(t *testing.T) {
	const nonExisting string = "iDoNotExist"
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Missing panic on non existing config file.")
		}
	}()

	InitClients(nonExisting)
}

// Tests that InitClients panics correctly, if the provided
// clients file is not a yaml file.
func TestInitClientsInvalidYaml(t *testing.T) {
	const invalidYaml string = "resources/fixtures/invalidYaml.txt"
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Missing panic on invalid yaml file.")
		}
	}()

	InitClients(invalidYaml)
}

// Tests that InitClients opens a proper clients yaml correctly.
func TestInitClientsYaml(t *testing.T) {
	const clientsYaml string = "resources/fixtures/testClients.yml"
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic trying to open clients yaml file '%s': '%v'\n", clientsYaml, r)
		}
	}()

	InitTestDb(t)

	InitClients(clientsYaml)
}
