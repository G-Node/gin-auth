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
	"io/ioutil"
	"time"

	"github.com/G-Node/gin-auth/util"
	"github.com/jmoiron/sqlx"
	"github.com/pborman/uuid"
	"gopkg.in/yaml.v2"
)

// Client object stored in the database
type Client struct {
	UUID             string
	Name             string
	Secret           string
	ScopeProvidedMap map[string]string
	ScopeWhitelist   util.StringSet
	ScopeBlacklist   util.StringSet
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

// listClientUUIDs returns a StringSet of the UUIDs of clients currently
// in the database.
func listClientUUIDs() util.StringSet {
	const q = "SELECT uuid FROM Clients"

	clients := make([]string, 0)
	err := database.Select(&clients, q)
	if err != nil {
		panic(err)
	}

	return util.NewStringSet(clients...)
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
	const q = `SELECT name FROM ClientScopeProvided`

	if scope.Len() == 0 {
		return false
	}

	check := []string{}
	err := database.Select(&check, q)
	if err != nil {
		panic(err)
	}

	global := util.NewStringSet(check...)
	return global.IsSuperset(scope)
}

// DescribeScope turns a scope into a map of names to descriptions.
// If the map is complete the second return value is true.
func DescribeScope(scope util.StringSet) (map[string]string, bool) {
	const q = `SELECT name, description FROM ClientScopeProvided`

	desc := make(map[string]string)
	if scope.Len() == 0 {
		return desc, false
	}

	data := []struct {
		Name        string
		Description string
	}{}

	err := database.Select(&data, q)
	if err != nil {
		panic(err)
	}

	names := make([]string, len(data))
	for i, d := range data {
		names[i] = d.Name
		if scope.Contains(d.Name) {
			desc[d.Name] = d.Description
		}
	}
	global := util.NewStringSet(names...)

	return desc, global.IsSuperset(scope)
}

// ScopeProvided the scope provided by this client as a StringSet.
// The scope is extracted from the clients ScopeProvidedMap.
func (client *Client) ScopeProvided() util.StringSet {
	scope := make([]string, 0, len(client.ScopeProvidedMap))
	for s := range client.ScopeProvidedMap {
		scope = append(scope, s)
	}
	return util.NewStringSet(scope...)
}

// ApprovalForAccount gets a client approval for this client which was
// approved for a specific account.
func (client *Client) ApprovalForAccount(accountUUID string) (*ClientApproval, bool) {
	const q = `SELECT * FROM ClientApprovals WHERE clientUUID = $1 AND accountUUID = $2`

	approval := &ClientApproval{}
	err := database.Get(approval, q, client.UUID, accountUUID)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	return approval, err == nil
}

// Approve creates a new client approval or extends an existing approval, such that the
// given scope is is approved for the given account.
func (client *Client) Approve(accountUUID string, scope util.StringSet) (err error) {
	if !CheckScope(scope) {
		return errors.New("Invalid scope")
	}

	if scope.Intersect(client.ScopeBlacklist).Len() > 0 {
		return errors.New("Blacklisted scope")
	}

	scope = scope.Difference(client.ScopeWhitelist)
	if scope.Len() == 0 {
		return nil
	}

	approval, ok := client.ApprovalForAccount(accountUUID)
	if ok {
		// approval exists
		if !approval.Scope.IsSuperset(scope) {
			approval.Scope = approval.Scope.Union(scope)
			err = approval.Update()
		}
	} else {
		// create new approval
		approval = &ClientApproval{
			ClientUUID:  client.UUID,
			AccountUUID: accountUUID,
			Scope:       scope,
		}
		err = approval.Create()
	}
	return err
}

// CreateGrantRequest check whether response type, redirect URI and scope are valid and creates a new
// grant request for this client. Grant types are defined by RFC6749 "OAuth 2.0 Authorization Framework"
// Supported grant types are: "code" (authorization code), "token" (implicit request),
// "owner" (resource owner password credentials), "client" (client credentials)
func (client *Client) CreateGrantRequest(responseType, redirectURI, state string, scope util.StringSet) (*GrantRequest, error) {
	if !(responseType == "code" || responseType == "token" || responseType == "owner" || responseType == "client") {
		return nil, errors.New("Response type expected to be one of the following: 'code', 'token', 'owner', 'client'")
	}
	if !client.RedirectURIs.Contains(redirectURI) {
		return nil, fmt.Errorf("Redirect URI invalid: '%s'", redirectURI)
	}
	if !CheckScope(scope) {
		return nil, errors.New("Invalid scope")
	}
	if scope.Intersect(client.ScopeBlacklist).Len() > 0 {
		return nil, errors.New("Blacklisted scope")
	}
	if state == "" {
		return nil, errors.New("Missing client state")
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

// delete removes a client from a database via a transaction.
func (client *Client) delete(tx *sqlx.Tx) error {
	const q = `DELETE FROM Clients c WHERE c.uuid=$1`

	_, err := tx.Exec(q, client.UUID)

	return err
}

// create stores a new client in the database.
func (client *Client) create(tx *sqlx.Tx) error {
	const q = `INSERT INTO Clients (uuid, name, secret, scopeWhitelist, scopeBlacklist, redirectURIs, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, $6, now(), now())
	           RETURNING *`
	const qScope = `INSERT INTO ClientScopeProvided (clientUUID, name, description)
	                VALUES ($1, $2, $3)`

	if client.UUID == "" {
		client.UUID = uuid.NewRandom().String()
	}

	err := tx.Get(client, q, client.UUID, client.Name, client.Secret, client.ScopeWhitelist,
		client.ScopeBlacklist, client.RedirectURIs)
	if err == nil {
		for k, v := range client.ScopeProvidedMap {
			_, err = tx.Exec(qScope, client.UUID, k, v)
			if err != nil {
				break
			}
		}
	}

	return err
}

// deleteScope removes all scopes corresponding to a client uuid from the database.
func (client *Client) deleteScope(tx *sqlx.Tx) error {
	const q = `DELETE FROM ClientScopeProvided WHERE clientuuid=$1`

	_, err := tx.Exec(q, client.UUID)

	return err
}

// createScope adds all client scopes from a Client to the database.
func (client *Client) createScope(tx *sqlx.Tx) error {
	const qScope = `INSERT INTO ClientScopeProvided (clientUUID, name, description)
	                VALUES ($1, $2, $3)`

	var err error
	for k, v := range client.ScopeProvidedMap {
		_, err = tx.Exec(qScope, client.UUID, k, v)
		if err != nil {
			break
		}
	}
	return err
}

// update removes all scopes associated with a specific Client from the database,
// updates all client database fields and adds new scopes with data from this Client.
func (client *Client) update(tx *sqlx.Tx) error {
	const q = `UPDATE Clients
	           SET name=$2, secret=$3, scopeWhitelist=$4, scopeBlacklist=$5, redirectURIs=$6, updatedAt=now()
	           WHERE uuid=$1`

	err := client.deleteScope(tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(q, client.UUID, client.Name, client.Secret, client.ScopeWhitelist,
		client.ScopeBlacklist, client.RedirectURIs)
	if err != nil {
		return err
	}

	if len(client.ScopeProvidedMap) > 0 {
		err = client.createScope(tx)
	}

	return err
}

// InitClients loads client information from a yaml configuration file
// and updates the corresponding entries in the database.
func InitClients(path string) {

	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	confClients := make([]struct {
		UUID           string            `yaml:"UUID"`
		Name           string            `yaml:"Name"`
		Secret         string            `yaml:"Secret"`
		ScopeProvided  map[string]string `yaml:"ScopeProvided"`
		ScopeWhitelist []string          `yaml:"ScopeWhitelist"`
		ScopeBlacklist []string          `yaml:"ScopeBlacklist"`
		RedirectURIs   []string          `yaml:"RedirectURIs"`
	}, 0)

	err = yaml.Unmarshal(content, &confClients)
	if err != nil {
		panic(err)
	}

	clients := make([]Client, len(confClients))
	for i, cl := range confClients {
		clients[i].UUID = cl.UUID
		clients[i].Name = cl.Name
		clients[i].Secret = cl.Secret
		clients[i].ScopeProvidedMap = cl.ScopeProvided
		clients[i].ScopeWhitelist = util.NewStringSet(cl.ScopeWhitelist...)
		clients[i].ScopeBlacklist = util.NewStringSet(cl.ScopeBlacklist...)
		clients[i].RedirectURIs = util.NewStringSet(cl.RedirectURIs...)
	}

	updateClients(clients)
}

// updateDatabase updates the clients and clientScopeProvided tables
// with the contents of []Client.
func updateClients(confClients []Client) {
	clientIDs := make([]string, len(confClients), len(confClients))
	for i, v := range confClients {
		clientIDs[i] = v.UUID
	}

	confClientIDs := util.NewStringSet(clientIDs...)
	dbClientIDs := listClientUUIDs()
	removeDbClients := dbClientIDs.Difference(confClientIDs)

	tx := database.MustBegin()

	var err error
	if len(removeDbClients) > 0 {
		for remID := range removeDbClients {
			remClient, clientExists := GetClient(remID)
			if clientExists {
				err = remClient.delete(tx)
				if err != nil {
					break
				}
			}
		}
	}

	for _, cl := range confClients {
		if dbClientIDs.Contains(cl.UUID) {
			err = cl.update(tx)
		} else {
			err = cl.create(tx)
		}
		if err != nil {
			break
		}
	}

	if err != nil {
		errTx := tx.Rollback()
		if errTx != nil {
			err = fmt.Errorf("After initial error '%v'\nrollback failed: '%v'\n", err, errTx)
		}
		panic(err)
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}
