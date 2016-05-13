// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
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
func RemoveExpired(grantLifeTime time.Duration) {
	const delGrant = `DELETE from GrantRequests WHERE createdAt < $1`
	database.MustExec(delGrant, time.Now().Add(-time.Minute*grantLifeTime))

	const q = `DELETE from AccessTokens WHERE expires < now();
		   DELETE from Sessions WHERE expires < now();`
	database.MustExec(q)
}

// RunCleaner starts an infinite loop which
// periodically executes the RemoveExpired function.
func RunCleaner(srvConf *conf.ServerConfig) {
	t := time.NewTicker(time.Minute * srvConf.CleanerInterval)
	for {
		select {
		case <-t.C:
			RemoveExpired(srvConf.GrantReqLifeTime)
		}
	}
}
