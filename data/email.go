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
