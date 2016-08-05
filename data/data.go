// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/G-Node/gin-auth/conf"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // pg driver needs to be imported in order to load it
)

var database *sqlx.DB

// InitDb initializes a global database connection.
// An existing connection will be closed.
func InitDb(config *conf.DbConfig) {
	if database != nil {
		database.Close()
	}

	var err error
	database, err = sqlx.Connect(config.Driver, config.Open)
	if err != nil {
		panic(err)
	}
}

// InitTestDb initializes a database for testing purpose.
func InitTestDb(t *testing.T) {
	config := conf.GetDbConfig()
	InitDb(config)

	fixtures, err := ioutil.ReadFile(conf.GetResourceFile("fixtures", "testdb.sql"))
	if err != nil {
		t.Fatal(err)
	}
	database.MustExec(string(fixtures))
}

// RemoveExpired removes rows of expired entries from
// AccessTokens, Sessions and GrantRequests database tables.
func RemoveExpired() {
	const delGrant = `DELETE from GrantRequests WHERE createdAt <= $1`
	database.MustExec(delGrant, time.Now().Add(-1*conf.GetServerConfig().GrantReqLifeTime))

	const q = `DELETE from AccessTokens WHERE expires <= now();
		   DELETE from Sessions WHERE expires <= now();`
	database.MustExec(q)
}

// RemoveStaleAccounts removes all accounts that where registered,
// but never accessed within a defined period of time
func RemoveStaleAccounts() {
	const q = `DELETE FROM Accounts WHERE
	 	   NOT isdisabled AND
	 	   resetpwcode IS NULL AND
	 	   activationcode IS NOT NULL AND
	 	   updatedat < $1`
	database.MustExec(q, time.Now().Add(-1*conf.GetServerConfig().UnusedAccountLifeTime))
}

// RunCleaner starts an infinite loop which
// periodically executes database cleanup functions.
func RunCleaner() {
	go func() {
		// TODO add log entry once logging is implemented
		t := time.NewTicker(conf.GetServerConfig().CleanerInterval)
		for {
			select {
			case <-t.C:
				RemoveExpired()
				RemoveStaleAccounts()
			}
		}
	}()
}

// EmailDispatch checks e-mail queue database entries, handles the entries
// according to the smtp mode setting and removes the entries after they successful handling.
func EmailDispatch() {
	emails, err := GetQueuedEmails()
	if err != nil {
		panic(err)
	}
	for _, email := range emails {
		err = email.Send()
		if err != nil {
			// TODO log if an error occurs trying to send an e-mail, but continue
			fmt.Printf("Error trying to send e-mail (Id %d): %s\n", email.Id, err.Error())
		} else {
			err = email.Delete()
			if err != nil {
				panic(err)
			}
		}
	}
}
