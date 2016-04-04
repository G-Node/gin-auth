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
	"time"
)

// OAuthClient object stored in the database
type OAuthClient struct {
	UUID          string
	Name          string
	Secret        string
	ScopeProvided SqlStringSlice
	RedirectURIs  SqlStringSlice
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ListOAuthClients returns all registered OAuth clients ordered by name
func ListOAuthClients() []OAuthClient {
	const q = `SELECT * FROM OAuthClients ORDER BY name`

	clients := make([]OAuthClient, 0)
	err := database.Select(&clients, q)
	if err != nil {
		panic(err)
	}

	return clients
}

// GetOAuthClient returns an OAuth client with a given uuid.
// Returns an error if no client with a matching uuid can be found.
func GetOAuthClient(uuid string) (*OAuthClient, error) {
	const q = `SELECT * FROM OAuthClient c WHERE c.uuid=$1`

	client := &OAuthClient{}
	err := database.Get(client, q, uuid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return client, err
}

// GetOAuthClientByName returns an OAuth client with a given client name.
// Returns an error if no client with a matching name can be found.
func GetOAuthClientByName(name string) (*OAuthClient, error) {
	const q = `SELECT * FROM OAuthClient c WHERE c.name=$1`

	client := &OAuthClient{}
	err := database.Get(client, q, name)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return client, err
}

// Create stores a new client in the database.
func (client *OAuthClient) Create() error {
	const q = `INSERT INTO OAuthClients (uuid, name, secret, scopeProvided, redirectURIs, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, now(), now())
	           RETURNING *`

	return database.Get(client, q, client.UUID, client.Name, client.Secret, client.ScopeProvided, client.RedirectURIs)
}

// Delete removes an existing client from the database
func (client *OAuthClient) Delete() error {
	const q = `DELETE FROM OAuthClients c WHERE c.uuid=$1`

	_, err := database.Exec(q, client.UUID)
	return err
}
