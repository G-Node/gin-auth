// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"database/sql"
	"time"

	"github.com/G-Node/gin-auth/util"
	"github.com/G-Node/gin-auth/conf"
)

// Email data as stored in the database
type Email struct {
	Id        int
	Mode      sql.NullString
	Sender    string
	Recipient util.StringSet
	Content   []byte
	CreatedAt time.Time
}

// GetQueuedEmails selects all unsent e-mails from the email queue
// database table and returns the result as a slice of Emails.
func GetQueuedEmails() ([]Email, error) {
	const q = `SELECT * FROM EmailQueue order by createdat`

	emails := make([]Email, 0)
	err := database.Select(&emails, q)

	return emails, err
}

// Create adds a new entry to table EmailQueue
func (e *Email) Create(to util.StringSet, content []byte) error {

	const q = `INSERT INTO EmailQueue(mode, sender, recipient, content, createdat)
	           VALUES ($1, $2, $3, $4, now())
	           RETURNING *`

	config := conf.GetSmtpCredentials()
	mode := sql.NullString{}
	mode.Scan(config.Mode)
	err := database.Get(e, q, mode, config.From, to, content)

	return err
}

// Delete removes the current e-mail from table EmailQueue
func (e *Email) Delete() error {
	const q = `DELETE FROM EmailQueue WHERE id=$1`
	_, err := database.Exec(q, e.Id)
	return err
}
