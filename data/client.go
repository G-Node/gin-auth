// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/G-Node/gin-auth/util"
	"github.com/pborman/uuid"
	"time"
)

// Client object stored in the database
type Client struct {
	UUID             string
	Name             string
	Secret           string
	ScopeProvidedMap map[string]string
	RedirectURIs     util.StringSet
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ListClients returns all registered OAuth clients ordered by name
func ListClients() []Client {
	const q = `SELECT * FROM Clients ORDER BY name`

	clients := make([]Client, 0)
	err := database.Select(&clients, q)
	if err != nil {
		panic(err)
	}

	return clients
}

// GetClient returns an OAuth client with a given uuid.
// Returns false if no client with a matching uuid can be found.
func GetClient(uuid string) (*Client, bool) {
	const q = `SELECT * FROM Clients WHERE uuid=$1`
	return getClient(q, uuid)
}

// GetClientByName returns an OAuth client with a given client name.
// Returns false if no client with a matching name can be found.
func GetClientByName(name string) (*Client, bool) {
	const q = `SELECT * FROM Clients WHERE name=$1`
	return getClient(q, name)
}

func getClient(q, parameter string) (*Client, bool) {
	const qScope = `SELECT name, description FROM ClientScopeProvided WHERE clientUUID = $1`

	client := &Client{ScopeProvidedMap: make(map[string]string)}
	err := database.Get(client, q, parameter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}

	scope := []struct {
		Name        string
		Description string
	}{}
	err = database.Select(&scope, qScope, client.UUID)
	if err != nil {
		panic(err)
	}
	for _, s := range scope {
		client.ScopeProvidedMap[s.Name] = s.Description
	}

	return client, true
}

// CheckScope checks whether a certain scope exists by searching
// through all provided scopes from all registered clients.
func CheckScope(scope util.StringSet) bool {
	const q = `SELECT name FROM ClientScopeProvided WHERE array_position($1, name) >= 0`

	if scope.Len() == 0 {
		return false
	}

	check := []string{}
	err := database.Select(&check, q, scope)
	if err != nil {
		panic(err)
	}

	global := util.NewStringSet(check...)
	return global.IsSuperset(scope)
}

// DescribeScope turns a scope into a map of names to descriptions.
// If the map is complete the second return value is true.
func DescribeScope(scope util.StringSet) (map[string]string, bool) {
	const q = `SELECT name, description FROM ClientScopeProvided WHERE array_position($1, name) >= 0`

	desc := make(map[string]string)
	if scope.Len() == 0 {
		return desc, false
	}

	data := []struct {
		Name        string
		Description string
	}{}

	err := database.Select(&data, q, scope)
	if err != nil {
		panic(err)
	}

	names := make([]string, len(data))
	for i, d := range data {
		names[i] = d.Name
		desc[d.Name] = d.Description
	}
	global := util.NewStringSet(names...)

	return desc, global.IsSuperset(scope)
}

func (client *Client) ScopeProvided() util.StringSet {
	scope := make([]string, len(client.ScopeProvidedMap))
	for s := range client.ScopeProvidedMap {
		scope = append(scope, s)
	}
	return util.NewStringSet(scope...)
}

// Create stores a new client in the database.
func (client *Client) Create() error {
	const q = `INSERT INTO Clients (uuid, name, secret, redirectURIs, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, now(), now())
	           RETURNING *`
	const qScope = `INSERT INTO ClientScopeProvided (clientUUID, name, description)
					VALUES ($1, $2, $3)`

	if client.UUID == "" {
		client.UUID = uuid.NewRandom().String()
	}

	tx := database.MustBegin()
	err := tx.Get(client, q, client.UUID, client.Name, client.Secret, client.RedirectURIs)
	if err != nil {
		tx.Rollback()
		return err
	}
	for k, v := range client.ScopeProvidedMap {
		_, err = tx.Exec(qScope, client.UUID, k, v)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (client *Client) CreateGrantRequest(responseType, redirectURI, state string, scope util.StringSet) (*GrantRequest, error) {
	if !(responseType == "code" || responseType == "token") {
		return nil, errors.New("Response type expected to be 'code' or 'token'")
	}
	if !client.RedirectURIs.Contains(redirectURI) {
		return nil, errors.New(fmt.Sprintf("Redirect URI invalid: '%s'", redirectURI))
	}
	if !CheckScope(scope) {
		return nil, errors.New("Invalid scope")
	}

	request := &GrantRequest{
		GrantType:      responseType,
		RedirectURI:    redirectURI,
		State:          state,
		ScopeRequested: scope,
		ClientUUID:     client.UUID}
	err := request.Create()

	return request, err
}

// Delete removes an existing client from the database
func (client *Client) Delete() error {
	const q = `DELETE FROM Clients c WHERE c.uuid=$1`

	_, err := database.Exec(q, client.UUID)
	return err
}
