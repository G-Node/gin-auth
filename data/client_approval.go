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
	"github.com/G-Node/gin-auth/util"
	"github.com/pborman/uuid"
	"time"
)

// ClientApproval contains information about scopes a user has already
// approved for a certain client. This is needed to implement Trust On
// First Use (TOFU).
type ClientApproval struct {
	UUID        string
	Scope       util.StringSet
	ClientUUID  string
	AccountUUID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListClientApprovals returns all client approvals stored in the database
// ordered by creation time.
func ListClientApprovals() []ClientApproval {
	const q = `SELECT * FROM ClientApprovals ORDER BY createdAt`

	approvals := make([]ClientApproval, 0)
	err := database.Select(&approvals, q)
	if err != nil {
		panic(err)
	}

	return approvals
}

// GetClientApproval retrieves an approval with a given UUID.
// Returns false if no matching approval exists.
func GetClientApproval(uuid string) (*ClientApproval, bool) {
	const q = `SELECT * FROM ClientApprovals WHERE uuid=$1`

	approval := &ClientApproval{}
	err := database.Get(approval, q, uuid)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return approval, err == nil
}

// Create stores a new approval in the database.
// If the UUID is empty a new random UUID will be created.
func (app *ClientApproval) Create() error {
	const q = `INSERT INTO ClientApprovals (uuid, scope, clientUUID, accountUUID, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, now(), now())
	           RETURNING *`

	if app.UUID == "" {
		app.UUID = uuid.NewRandom().String()
	}

	return database.Get(app, q, app.UUID, app.Scope, app.ClientUUID, app.AccountUUID)
}

// Update stores the new values of the approval in the database.
// New values for CreatedAt will be ignored. UpdatedAt will be set
// automatically to the current time.
func (app *ClientApproval) Update() error {
	const q = `UPDATE ClientApprovals SET (scope, clientUUID, accountUUID, updatedAt) = ($1, $2, $3, now())
	           WHERE uuid=$4
	           RETURNING *`

	return database.Get(app, q, app.Scope, app.ClientUUID, app.AccountUUID, app.UUID)
}

// Delete removes an approval from the database.
func (app *ClientApproval) Delete() error {
	const q = `DELETE FROM ClientApprovals WHERE uuid=$1`

	_, err := database.Exec(q, app.UUID)
	return err
}
