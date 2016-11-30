// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"strings"
	"testing"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/util"
	"github.com/pborman/uuid"
)

const (
	uuidClientGin = "8b14d6bb-cae7-4163-bbd1-f3be46e43e31"
	uuidClientWB  = "177c56a4-57b4-4baf-a1a7-04f3d8e5b276"
)

func TestListClients(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	clients := ListClients()
	if len(clients) != 2 {
		t.Error("Expected number of clients does not match")
	}
}

func TestListClientUUIDs(t *testing.T) {
	defer util.FailOnPanic(t)
	InitTestDb(t)

	clientList := listClientUUIDs()
	if len(clientList) != 2 || !clientList.Contains(uuidClientGin) {
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
	if len(desc) != 2 {
		t.Error("DescribeScope returned false number of scopes.")
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

func TestClient_ApprovalForAccount_Approve(t *testing.T) {
	InitTestDb(t)

	client, ok := GetClient(uuidClientGin)
	if !ok {
		t.Error("Client does not exist")
	}

	// get existing approval
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

	// get non existing approval
	_, ok = client.ApprovalForAccount(uuidBob)
	if ok {
		t.Error("Approval should not exist")
	}

	// approve a new client
	err := client.Approve(uuidBob, util.NewStringSet("repo-read"))
	if err != nil {
		t.Error(err)
	}
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

	// expand approval for a client
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

	// client "wb"
	// whitelist: account-read, repo-read
	// blacklist: account-admin
	client, ok = GetClient(uuidClientWB)

	// approve whitelisted scope
	err = client.Approve(uuidBob, util.NewStringSet("repo-read", "account-read"))
	if err != nil {
		t.Error("Approving a whitelisted scope should not result in an error")
	}
	_, ok = client.ApprovalForAccount(uuidBob)
	if ok {
		t.Error("Approving a whitelisted scope should not create an approval")
	}

	// approve blacklisted scope
	err = client.Approve(uuidBob, util.NewStringSet("account-admin"))
	if err == nil {
		t.Error("Approving a blacklisted scope should result in an error")
	}
	_, ok = client.ApprovalForAccount(uuidBob)
	if ok {
		t.Error("Approving a blacklisted scope should not create an approval")
	}

	// approve partially whitelisted scope
	err = client.Approve(uuidBob, util.NewStringSet("account-read", "account-write"))
	if err != nil {
		t.Error("Approving a partially whitelisted scope should not result in an error")
	}
	approval, ok = client.ApprovalForAccount(uuidBob)
	if !ok {
		t.Error("Approving a partially whitelisted scope must create an approval")
	}
	if approval.Scope.Contains("account-read") {
		t.Error("Scope should not contain 'account-read' (whitelisted)")
	}
	if !approval.Scope.Contains("account-write") {
		t.Error("Scope should contain 'account-write'")
	}
}

func TestClient_CreateGrantRequest(t *testing.T) {
	InitTestDb(t)

	client, ok := GetClient(uuidClientGin)
	if !ok {
		t.Error("Client does not exist")
	}

	validResponseType := "code"
	validState := util.RandomToken()
	validRedirectURI := client.RedirectURIs.Strings()[0]
	validScope := util.NewStringSet("repo-read")

	// Test invalid response type
	_, err := client.CreateGrantRequest("foo", validRedirectURI, validState, validScope)
	if err == nil || !strings.Contains(err.Error(), "Response type expected") {
		t.Error("Error expected")
	}

	// Test invalid redirect
	_, err = client.CreateGrantRequest(validResponseType, "https://doesnotexist.com/callback", validState, validScope)
	if err == nil || !strings.Contains(err.Error(), "Redirect URI invalid") {
		t.Error("Error expected")
	}

	// Test invalid scope
	_, err = client.CreateGrantRequest(validResponseType, validRedirectURI, validState, util.NewStringSet("foo-read"))
	if err == nil || !strings.Contains(err.Error(), "Invalid scope") {
		t.Error("Error expected")
	}

	// Test blacklisted scope
	_, err = client.CreateGrantRequest(validResponseType, validRedirectURI, validState, util.NewStringSet("account-admin"))
	if err == nil || !strings.Contains(err.Error(), "Blacklisted scope") {
		t.Error("Error expected")
	}

	// Test missing client state token
	_, err = client.CreateGrantRequest(validResponseType, validRedirectURI, "", validScope)
	if err == nil || !strings.Contains(err.Error(), "Missing client state") {
		t.Error("Error expected")
	}

	// all OK
	request, err := client.CreateGrantRequest(validResponseType, validRedirectURI, validState, validScope)
	if err != nil {
		t.Error(err)
	}
	if request.ClientUUID != client.UUID {
		t.Error("Client UUID does not match")
	}
	if !request.ScopeRequested.Contains("repo-read") {
		t.Error("The requested scope should contain 'repo-read'")
	}
	if request.State != validState {
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
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Missing panic on invalid yaml file.")
		}
	}()
	InitClients(conf.GetResourceFile("fixtures", "invalidYaml.txt"))
}

// Tests the insertion of a client into the database.
func TestClient_create(t *testing.T) {
	InitTestDb(t)

	const (
		testScope = "testEntry"
		testUri   = "https://testRedirecturi.com/somewhere"
	)

	var client = new(Client)
	client.UUID = uuid.NewRandom().String()
	client.Name = "TestClient" + client.UUID
	client.Secret = "TestSecret"
	client.ScopeProvidedMap = map[string]string{testScope: testScope}
	client.RedirectURIs = util.NewStringSet(testUri)

	tx := database.MustBegin()

	err := client.create(tx)
	if err != nil {
		t.Errorf("Error creating client '%s': '%v'", client.UUID, err)
	}
	tx.Commit()

	check, ok := GetClient(client.UUID)
	if !ok {
		t.Errorf("Client not created.")
	}

	if check.Name != client.Name {
		t.Errorf("DB client name '%s' does not match expected name '%s'",
			check.Name, client.Name)
	}
	if check.Secret != client.Secret {
		t.Errorf("DB client secret '%s' does not match expected secret '%s'",
			check.Secret, client.Secret)
	}
	if len(check.ScopeProvidedMap) != len(client.ScopeProvidedMap) {
		t.Errorf("Number of DB scope entries (%d) differ from expected entries (%d)",
			len(check.ScopeProvidedMap), len(client.ScopeProvidedMap))
	}
	if !check.ScopeProvided().Contains(testScope) {
		t.Errorf("DB scope entry does not contain expected scope entry '%s'", testScope)
	}
	if check.RedirectURIs.Len() != client.RedirectURIs.Len() {
		t.Errorf("Number of DB redirectURI entries (%d) differ from expected entries (%d)",
			check.RedirectURIs.Len(), client.RedirectURIs.Len())
	}
	if !check.RedirectURIs.Contains(testUri) {
		t.Errorf("DB redirectURI '%v' entry does not contain expected entry '%s'",
			check.RedirectURIs, testUri)
	}
}

// Tests various correct fails when trying to insert a client into the database.
func TestClient_createFail(t *testing.T) {
	InitTestDb(t)

	const (
		testScope = "testEntry"
		testUri   = "https://testRedirecturi.com/somewhere"
	)

	var client = new(Client)
	client.UUID = uuid.NewRandom().String()
	client.Name = "TestClient" + client.UUID
	client.Secret = "TestSecret"
	client.RedirectURIs = util.NewStringSet(testUri)
	client.ScopeProvidedMap = map[string]string{testScope: testScope}

	tx := database.MustBegin()
	err := client.create(tx)
	if err != nil {
		t.Errorf("Error creating client '%s': '%v'", client.UUID, err)
	}
	tx.Commit()

	// Test fail on incorrect uuid length
	tx = database.MustBegin()
	client.UUID = "1"
	err = client.create(tx)
	if err == nil {
		t.Errorf("Missing error on invalid UUID length: %v", client)
	}
	tx.Rollback()

	// Test fail on incorrect name length
	tx = database.MustBegin()
	client.UUID = uuid.NewRandom().String()
	client.Name = ""
	err = client.create(tx)
	if err == nil {
		t.Errorf("Missing error on invalid name length: %v", client)
	}

	// Test fail on duplicate name entry
	tx = database.MustBegin()
	client.UUID = uuid.NewRandom().String()
	err = client.create(tx)
	if err == nil {
		t.Error("Missing error on duplicate name.")
	}
	tx.Rollback()

	// Test fail duplicate client scope
	tx = database.MustBegin()
	client.Name = "TestClient" + client.UUID
	err = client.create(tx)
	if err == nil {
		t.Error("Missing error on duplicate client scope.")
	}
	tx.Rollback()
}

// Tests removal of a client and all of its scopes
// from the corresponding database tables.
func TestClient_delete(t *testing.T) {
	InitTestDb(t)

	const (
		testScope = "testEntry"
		testUri   = "https://testRedirecturi.com/somewhere"
	)

	var client = new(Client)
	client.UUID = uuid.NewRandom().String()
	client.Name = "TestClient" + client.UUID
	client.Secret = "TestSecret"
	client.RedirectURIs = util.NewStringSet(testUri)
	client.ScopeProvidedMap = map[string]string{testScope: testScope}

	tx := database.MustBegin()
	err := client.create(tx)
	if err != nil {
		t.Errorf("Error creating client '%s': '%v'", client.UUID, err)
	}
	tx.Commit()

	_, ok := GetClient(client.UUID)
	if !ok {
		t.Errorf("Client not created.")
	}

	if !CheckScope(util.NewStringSet(testScope)) {
		t.Error("Client scope not created.")
	}

	tx = database.MustBegin()
	err = client.delete(tx)
	if err != nil {
		t.Errorf("Error deleting client: %v", err)
	}
	tx.Commit()

	_, ok = GetClient(client.UUID)
	if ok {
		t.Errorf("Client not deleted.")
	}

	if CheckScope(util.NewStringSet(testScope)) {
		t.Error("Client scope was not deleted.")
	}
}

// Tests update of a client and proper update of its scopes in the
// corresponding database tables.
func TestClient_update(t *testing.T) {
	InitTestDb(t)

	const (
		scopeOne   = "testScope1"
		scopeTwo   = "testScope2"
		scopeThree = "testScope3"
		testUri    = "https://testRedirecturi.com/somewhere"
		testUriNew = "https://testRedirecturi.com/somewhere/else"
	)

	var client = new(Client)
	client.UUID = uuid.NewRandom().String()
	client.Name = "TestClient" + client.UUID
	client.Secret = "TestSecret"
	client.RedirectURIs = util.NewStringSet(testUri)
	client.ScopeProvidedMap = map[string]string{scopeOne: scopeOne}

	tx := database.MustBegin()
	err := client.create(tx)
	if err != nil {
		t.Errorf("Error creating client '%s': '%v'", client.UUID, err)
	}
	tx.Commit()

	var clUpdate = new(Client)
	clUpdate.UUID = client.UUID
	clUpdate.Name = "TestClient_up" + client.UUID
	clUpdate.Secret = "TestSecret_up"
	clUpdate.RedirectURIs = util.NewStringSet(testUriNew)
	clUpdate.ScopeProvidedMap = map[string]string{scopeTwo: scopeTwo, scopeThree: scopeThree}

	tx = database.MustBegin()
	err = clUpdate.update(tx)
	if err != nil {
		t.Error(err)
	}
	tx.Commit()

	check, ok := GetClient(clUpdate.UUID)
	if !ok {
		t.Errorf("Error retrieving client '%s'", clUpdate.UUID)
	}

	if check.Name != clUpdate.Name {
		t.Errorf("DB client name '%s' does not match expected '%s'",
			check.Name, clUpdate.Name)
	}
	if check.Secret != clUpdate.Secret {
		t.Errorf("DB client secret '%s' does not match expected '%s'",
			check.Secret, clUpdate.Secret)
	}
	if check.RedirectURIs.Len() != clUpdate.RedirectURIs.Len() {
		t.Errorf("Number of DB redirectURI entries (%d) differ from expected entries (%d)",
			check.RedirectURIs.Len(), client.RedirectURIs.Len())
	}
	if !check.RedirectURIs.Contains(testUriNew) {
		t.Errorf("DB redirectURI '%v' entry does not contain expected entry '%s'",
			check.RedirectURIs, testUriNew)
	}
	if len(check.ScopeProvidedMap) != len(clUpdate.ScopeProvidedMap) {
		t.Errorf("Number of DB scope entries (%d) differ from expected entries (%d)",
			len(check.ScopeProvidedMap), len(clUpdate.ScopeProvidedMap))
	}
	if !check.UpdatedAt.After(check.CreatedAt) {
		t.Error("TestClient_update: Field updatedAt was not properly updated.")
	}

	if CheckScope(util.NewStringSet(scopeOne)) {
		t.Errorf("Scope '%s' was not removed from DB.", scopeOne)
	}
	if !CheckScope(util.NewStringSet(scopeTwo)) || !CheckScope(util.NewStringSet(scopeThree)) {
		t.Errorf("Scopes were not properly updated.")
	}
}

// Tests correct insertion, update and removal of clients of the updateClients function.
func TestClient_updateClients(t *testing.T) {
	InitTestDb(t)

	const (
		scopeOne      = "testScope1"
		testUri       = "https://testRedirecturi.com/somewhere"
		testUriUpdate = "https://testRedirecturi.com/somewhere/else"
	)

	dbClients := ListClients()

	testClient := new(Client)
	testClient.UUID = uuid.NewRandom().String()
	testClient.Name = "TestClient"
	testClient.Secret = "TestSecret"
	testClient.RedirectURIs = util.NewStringSet(testUri)
	testClient.ScopeProvidedMap = map[string]string{scopeOne: scopeOne}

	clients := make([]Client, 0)
	clients = append(clients, dbClients...)
	clients = append(clients, *testClient)

	initClientNum := len(listClientUUIDs())

	updateClients(clients)

	insertClientNum := len(listClientUUIDs())

	_, ok := GetClient(testClient.UUID)
	if !ok {
		t.Error("Client was not created.")
	}
	if insertClientNum != initClientNum+1 {
		t.Error("Number of clients expected to increase by one after update.")
	}

	testClient.Secret = "AnotherTestSecret"
	testClient.RedirectURIs = util.NewStringSet(testUriUpdate)

	clients = make([]Client, 0)
	clients = append(clients, dbClients...)
	clients = append(clients, *testClient)

	updateClients(clients)

	updateClientNum := len(listClientUUIDs())

	testClient, ok = GetClient(testClient.UUID)
	if !ok {
		t.Error("Client was not created.")
	}
	if testClient.Secret != "AnotherTestSecret" {
		t.Error("Secret was not updated")
	}
	if testClient.RedirectURIs.Len() != 1 || !testClient.RedirectURIs.Contains(testUriUpdate) {
		t.Error("RedirectURI was not updated")
	}
	if insertClientNum != updateClientNum {
		t.Error("Number of clients expected not to change after update.")
	}

	clients = make([]Client, 0)
	clients = append(clients, dbClients...)

	updateClients(clients)

	remClientNum := len(listClientUUIDs())

	_, ok = GetClient(testClient.UUID)
	if ok {
		t.Errorf("Client '%s' was not properly deleted.", testClient.UUID)
	}
	if remClientNum != updateClientNum-1 {
		t.Error("Number of clients expected to increase by one after update.")
	}
}

// Tests that a failing client insert does a proper rollback before raising panic.
func TestClient_updateClientsFailInsert(t *testing.T) {
	InitTestDb(t)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Missing panic on false insert.")
		}
	}()

	dbClient, ok := GetClient(uuidClientGin)
	if !ok {
		t.Errorf("Client '%s' not found.", uuidClientGin)
	}

	addClient := new(Client)
	addClient.UUID = uuid.NewRandom().String()
	addClient.Name = "TestClient" + addClient.UUID
	addClient.Secret = "TestSecret"
	addClient.RedirectURIs = util.NewStringSet("https://uri.com/toNowhere")
	addClient.ScopeProvidedMap = map[string]string{"entry1": "entry1"}

	failClient := new(Client)
	failClient.UUID = uuid.NewRandom().String()
	failClient.Name = "gin"

	clients := make([]Client, 0)
	clients = append(clients, *dbClient, *addClient, *failClient)

	initClientNum := len(listClientUUIDs())

	updateClients(clients)

	insertClientNum := len(listClientUUIDs())

	_, ok = GetClient(failClient.UUID)
	if ok {
		t.Error("Client should not have been created.")
	}
	if initClientNum != insertClientNum {
		t.Error("Number of clients does not match expected number.")
	}
}

// Tests that a failing client update does a proper rollback before raising panic.
func TestClient_updateClientsFailUpdate(t *testing.T) {
	InitTestDb(t)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Missing panic on false update.")
		}
	}()

	dbClient, ok := GetClient(uuidClientGin)
	if !ok {
		t.Errorf("Client '%s' not found.", uuidClientGin)
	}

	addClient := new(Client)
	addClient.UUID = uuid.NewRandom().String()
	addClient.Name = "TestAddClient" + addClient.UUID
	addClient.RedirectURIs = util.NewStringSet("https://uri.com/toNowhere")
	addClient.ScopeProvidedMap = map[string]string{"entry1": "entry1"}

	failClient := new(Client)
	failClient.UUID = uuid.NewRandom().String()
	failClient.Name = "TestFailClient" + failClient.UUID
	failClient.RedirectURIs = util.NewStringSet("https://uri.com/toNowhere")
	failClient.ScopeProvidedMap = map[string]string{"entry2": "entry2"}

	clients := make([]Client, 0)
	clients = append(clients, *dbClient, *addClient, *failClient)

	updateClients(clients)

	insertClientNum := len(listClientUUIDs())

	failClient.Name = "gin"
	failClients := make([]Client, 0)
	failClients = append(failClients, *dbClient, *failClient)

	updateClients(failClients)

	failClientNum := len(listClientUUIDs())

	check, ok := GetClient(failClient.UUID)
	if !ok {
		t.Error("Update fail client is missing.")
	}
	if check.Name == failClient.Name {
		t.Error("Client name should not have been updated.")
	}
	_, ok = GetClient(addClient.UUID)
	if !ok {
		t.Error("Client should not have been deleted.")
	}
	if failClientNum != insertClientNum {
		t.Error("Number of clients does not match expected number.")
	}
}
