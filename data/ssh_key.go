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

// SSHKey object stored in the database.
type SSHKey struct {
	Fingerprint string
	Key         string
	Description string
	AccountUUID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListSSHKeys returns all stored ssh keys.
func ListSSHKeys() []SSHKey {
	const q = `SELECT * FROM SSHKeys ORDER BY fingerprint`

	keys := make([]SSHKey, 0)
	err := database.Select(&keys, q)
	if err != nil {
		panic(err)
	}

	return keys
}

// GetSSHKey returns an SSH key for a given fingerprint.
// Returns an error if no key with the fingerprint can be found.
func GetSSHKey(fingerprint string) (*SSHKey, error) {
	const q = `SELECT * FROM SSHKeys k WHERE k.fingerprint=$1`

	key := &SSHKey{}
	err := database.Get(key, q, fingerprint)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return key, err
}

// Create stores a new SSH key in the database.
func (key *SSHKey) Create() error {
	const q = `INSERT INTO SSHKeys (fingerprint, key, description, accountUUID, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, now(), now())
	           RETURNING *`

	return database.Get(key, q, key.Fingerprint, key.Key, key.Description, key.AccountUUID)
}

// Delete removes an existing SSH key from the database.
func (key *SSHKey) Delete() error {
	const q = `DELETE FROM SSHKeys k WHERE k.fingerprint=$1`

	_, err := database.Exec(q, key.Fingerprint)
	return err
}
