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
	"github.com/G-Node/gin-auth/util"
	"time"
)

// GrantRequest contains data about an ongoing authorization grant request.
type GrantRequest struct {
	Token          string
	GrantType      string
	State          string
	Code           sql.NullString
	ScopeRequested util.StringSet
	ScopeApproved  util.StringSet
	RedirectURI    string
	ClientUUID     string
	AccountUUID    sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ListGrantRequests returns all current grant requests ordered by creation time.
func ListGrantRequests() []GrantRequest {
	const q = `SELECT * FROM GrantRequests ORDER BY createdAt`

	grantRequests := make([]GrantRequest, 0)
	err := database.Select(&grantRequests, q)
	if err != nil {
		panic(err)
	}

	return grantRequests
}

// GetGrantRequest returns a grant request with a given token.
// Returns false if no request with a matching token exists.
func GetGrantRequest(token string) (*GrantRequest, bool) {
	const q = `SELECT * FROM GrantRequests WHERE token=$1`

	grantRequest := &GrantRequest{}
	err := database.Get(grantRequest, q, token)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return grantRequest, err == nil
}

// GetGrantRequestByCode returns a grant request with a given code.
// Returns false if no request with a matching code exists.
func GetGrantRequestByCode(code string) (*GrantRequest, bool) {
	const q = `SELECT * FROM GrantRequests WHERE code=$1 AND code IS NOT NULL`

	grantRequest := &GrantRequest{}
	err := database.Get(grantRequest, q, code)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return grantRequest, err == nil
}

// ClearOldGrantRequests removes requests older than 15 minutes
// and returns the number of removed requests.
func ClearOldGrantRequests() int64 {
	const q = `DELETE FROM GrantRequests WHERE createdAt < $1`

	minutesAgo15 := time.Now().Add(-time.Minute * 15)

	res, err := database.Exec(q, minutesAgo15)
	if err != nil {
		panic(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	return rows
}

// ExchangeCodeForTokens creates an access token and a refresh token.
// Finally the grant request will be deleted from the database, even if the token creation fails!
func (req *GrantRequest) ExchangeCodeForTokens() (string, string, error) {
	defer req.Delete()

	const qCreateRefresh = `INSERT INTO RefreshTokens (token, scope, clientUUID, accountUUID, createdAt, updatedAt)
	                        VALUES ($1, $2, $3, $4, now(), now())
	                        RETURNING *`
	const qCreateAccess = `INSERT INTO AccessTokens (token, scope, expires, clientUUID, accountUUID, createdAt, updatedAt)
	                        VALUES ($1, $2, $3, $4, $5, now(), now())
	                        RETURNING *`

	if !req.AccountUUID.Valid || !req.IsApproved() {
		return "", "", errors.New("Invalid grant request")
	}

	refresh := &RefreshToken{
		Token:       util.RandomToken(),
		Scope:       req.ScopeApproved,
		ClientUUID:  req.ClientUUID,
		AccountUUID: req.AccountUUID.String}
	access := &AccessToken{
		Token:       util.RandomToken(),
		Scope:       req.ScopeApproved,
		Expires:     time.Now().Add(DefaultTokenLifeTime),
		ClientUUID:  req.ClientUUID,
		AccountUUID: req.AccountUUID.String}

	tx := database.MustBegin()
	err := tx.Get(refresh, qCreateRefresh, refresh.Token, refresh.Scope, refresh.ClientUUID, refresh.AccountUUID)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}
	err = tx.Get(access, qCreateAccess, access.Token, access.Scope, access.Expires, access.ClientUUID, access.AccountUUID)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	return access.Token, refresh.Token, nil
}

// Create stores a new grant request.
func (req *GrantRequest) Create() error {
	const q = `INSERT INTO GrantRequests (token, grantType, state, code, scopeRequested, scopeApproved, redirectUri,
	                                      clientUUID, accountUUID, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
	           RETURNING *`

	if req.Token == "" {
		req.Token = util.RandomToken()
	}

	return database.Get(req, q, req.Token, req.GrantType, req.State, req.Code, req.ScopeRequested, req.ScopeApproved,
		req.RedirectURI, req.ClientUUID, req.AccountUUID)
}

// Update an existing grant request.
func (req *GrantRequest) Update() error {
	const q = `UPDATE GrantRequests gr
	           SET (grantType, state, code, scopeRequested, scopeApproved, redirectUri, clientUUID, accountUUID, updatedAt) =
	               ($1, $2, $3, $4, $5, $6, $7, $8, now())
	           WHERE token=$9
	           RETURNING *`

	return database.Get(req, q, req.GrantType, req.State, req.Code, req.ScopeRequested, req.ScopeApproved, req.RedirectURI,
		req.ClientUUID, req.AccountUUID, req.Token)
}

// Delete removes an existing request from the database.
func (req *GrantRequest) Delete() error {
	const q = `DELETE FROM GrantRequests WHERE token=$1`

	_, err := database.Exec(q, req.Token)
	return err
}

// GetClientApproval gets a matching client approval from the database
func (req *GrantRequest) GetClientApproval() (*ClientApproval, bool) {
	const q = `SELECT * FROM ClientApprovals
	           WHERE clientUUID=$1 AND accountUUID=$2`

	approval := &ClientApproval{}
	err := database.Get(approval, q, req.ClientUUID, req.AccountUUID)

	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return approval, err == nil
}

// ApproveScopes checks for a matching client approval, approves the requested
// scope and saves the result.
func (req *GrantRequest) ApproveScopes() bool {
	approval, ok := req.GetClientApproval()
	if !ok {
		return false
	}

	approved := approval.Scope.IsSuperset(req.ScopeRequested)
	if !approved {
		return false
	}

	req.ScopeApproved = req.ScopeRequested
	err := req.Update()
	return err == nil
}

// IsApproved just looks up whether the requested scope is in
// the approved scope.
func (req *GrantRequest) IsApproved() bool {
	if req.ScopeRequested.Len() == 0 || req.ScopeApproved.Len() == 0 {
		return false
	}
	return req.ScopeApproved.IsSuperset(req.ScopeRequested)
}
