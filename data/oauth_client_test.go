package data

import (
	"database/sql"
	"github.com/pborman/uuid"
	"testing"
)

const (
	uuidClientGin = "8b14d6bb-cae7-4163-bbd1-f3be46e43e31"
)

func TestListOAuthClients(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	clients := ListOAuthClients()
	if len(clients) != 1 {
		t.Error("Exactly one client expected in list")
	}
}

func TestGetOAuthClient(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	client, err := GetOAuthClient(uuidClientGin)
	if err != nil {
		t.Error(err)
	}
	if client.Name != "gin" {
		t.Error("Client name was expected to be 'gin'")
	}

	_, err = GetOAuthClient("doesNotExist")
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

func TestGetOAuthClientByName(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	initTestDb(t)

	client, err := GetOAuthClientByName("gin")
	if err != nil {
		t.Error(err)
	}
	if client.UUID != uuidClientGin {
		t.Errorf("Client UUID was expected to be '%s'", uuidClientGin)
	}

	_, err = GetOAuthClientByName("doesNotExist")
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}

func TestCreateOAuthClient(t *testing.T) {
	initTestDb(t)

	id := uuid.NewRandom().String()
	new := OAuthClient{
		UUID:          id,
		Name:          "gin-foo",
		Secret:        "secret",
		ScopeProvided: SqlStringSlice{"foo-read", "foo-write"},
		RedirectURIs:  SqlStringSlice{"https://foo.com/redirect"}}

	err := new.Create()
	if err != nil {
		t.Error(err)
	}

	check, err := GetOAuthClient(id)
	if err != nil {
		t.Error(err)
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

func TestDeleteOAuthClient(t *testing.T) {
	initTestDb(t)

	client, err := GetOAuthClient(uuidClientGin)
	if err != nil {
		t.Error(err)
	}

	err = client.Delete()
	if err != nil {
		t.Error(err)
	}

	_, err = GetOAuthClient(uuidClientGin)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error("Error must be sql.ErrNoRows")
		}
	} else {
		t.Error("Error expected")
	}
}
