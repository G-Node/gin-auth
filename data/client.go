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
	"github.com/pborman/uuid"
	"time"
)

// Client object stored in the database
type Client struct {
	UUID          string
	Name          string
	Secret        string
	ScopeProvided SqlStringSlice
	RedirectURIs  SqlStringSlice
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
	const q = `SELECT * FROM Clients c WHERE c.uuid=$1`

	client := &Client{}
	err := database.Get(client, q, uuid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return client, err == nil
}

// GetClientByName returns an OAuth client with a given client name.
// Returns false if no client with a matching name can be found.
func GetClientByName(name string) (*Client, bool) {
	const q = `SELECT * FROM Clients c WHERE c.name=$1`

	client := &Client{}
	err := database.Get(client, q, name)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return client, err == nil
}

// Create stores a new client in the database.
func (client *Client) Create() error {
	const q = `INSERT INTO Clients (uuid, name, secret, scopeProvided, redirectURIs, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, now(), now())
	           RETURNING *`

	if client.UUID == "" {
		client.UUID = uuid.NewRandom().String()
	}

	return database.Get(client, q, client.UUID, client.Name, client.Secret, client.ScopeProvided, client.RedirectURIs)
}

// Delete removes an existing client from the database
func (client *Client) Delete() error {
	const q = `DELETE FROM Clients c WHERE c.uuid=$1`

	_, err := database.Exec(q, client.UUID)
	return err
}
