package data

import (
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

}

func TestGetOAuthClientByName(t *testing.T) {

}

func TestCreateOAuthClient(t *testing.T) {

}

func TestDeleteOAuthClient(t *testing.T) {

}
