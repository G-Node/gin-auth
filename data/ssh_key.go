// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/G-Node/gin-auth/conf"
	"github.com/G-Node/gin-auth/util"
	"github.com/G-Node/gin-core/gin"
	"golang.org/x/crypto/ssh"
)

// SSHKey object stored in the database.
type SSHKey struct {
	Fingerprint string
	Key         string
	Description string
	AccountUUID string
	Temporary   bool
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

// GetSSHKey returns an SSH key (permanent or temporary) for a given fingerprint.
// Returns false if no permanent key with the fingerprint can be found.
// Returns false if no temporary key with the fingerprint created within the LifeTime of temporary ssh keys can be found.
func GetSSHKey(fingerprint string) (*SSHKey, bool) {
	const q = `SELECT * FROM SSHKeys k WHERE k.fingerprint=$1 AND (NOT temporary OR createdat > $2)`

	key := &SSHKey{}
	err := database.Get(key, q, fingerprint, time.Now().Add(-1*conf.GetServerConfig().TmpSshKeyLifeTime))
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	return key, err == nil
}

// Create stores a new SSH key in the database.
func (key *SSHKey) Create() error {
	const q = `INSERT INTO SSHKeys (fingerprint, key, description, accountUUID, temporary, createdAt, updatedAt)
	           VALUES ($1, $2, $3, $4, $5, now(), now())
	           RETURNING *`

	return database.Get(key, q, key.Fingerprint, key.Key, key.Description, key.AccountUUID, key.Temporary)
}

// Delete removes an existing SSH key from the database.
func (key *SSHKey) Delete() error {
	const q = `DELETE FROM SSHKeys k WHERE k.fingerprint=$1`

	_, err := database.Exec(q, key.Fingerprint)
	return err
}

// SSHKeyMarshaler wraps a SSHKey together with an Account to provide all
// information needed to marshal a SSHKey
type SSHKeyMarshaler struct {
	SSHKey  *SSHKey
	Account *Account
}

// MarshalJSON implements Marshaler for SSHKeyMarshaler
func (keyMarshaler *SSHKeyMarshaler) MarshalJSON() ([]byte, error) {
	jsonData := gin.SSHKey{
		URL:         conf.MakeUrl("/api/keys?fingerprint=%s", keyMarshaler.SSHKey.Fingerprint),
		Fingerprint: keyMarshaler.SSHKey.Fingerprint,
		Key:         keyMarshaler.SSHKey.Key,
		Description: keyMarshaler.SSHKey.Description,
		Login:       keyMarshaler.Account.Login,
		AccountURL:  conf.MakeUrl("/api/accounts/%s", keyMarshaler.Account.Login),
		CreatedAt:   keyMarshaler.SSHKey.CreatedAt,
		UpdatedAt:   keyMarshaler.SSHKey.UpdatedAt,
	}
	return json.Marshal(jsonData)
}

// UnmarshalJSON implements Unmarshaler for Account.
// Only parses updatable fields: Key and Description.
// The fingerprint is parsed from the key.
func (key *SSHKey) UnmarshalJSON(bytes []byte) error {
	jsonData := &struct {
		Key         string `json:"key"`
		Description string `json:"description"`
		Temporary   bool   `json:"temporary"`
	}{}
	err := json.Unmarshal(bytes, jsonData)
	if err != nil {
		return err
	}

	parsed, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(jsonData.Key))
	if err != nil {
		return &util.ValidationError{
			Message:     "Unable to process key",
			FieldErrors: map[string]string{"key": "Invalid key"},
		}
	}

	sha := sha256.New()
	_, err = sha.Write(parsed.Marshal())
	if err != nil {
		panic(err)
	}
	fingerprint := base64.StdEncoding.EncodeToString(sha.Sum(nil))

	key.Key = jsonData.Key

	// Remove any base64 "=" padding characters
	key.Fingerprint = strings.TrimRight(fingerprint, "=")
	if jsonData.Description != "" {
		key.Description = jsonData.Description
	} else {
		key.Description = comment
	}

	return nil
}
