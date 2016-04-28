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

	"github.com/G-Node/gin-auth/conf"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // pg driver needs to be imported in order to load it
)

var database *sqlx.DB

// InitDb initializes a global database connection.
// An existing connection will be closed.
func InitDb(config *conf.DbConfig) (err error) {
	if database != nil {
		database.Close()
	}
	database, err = sqlx.Connect(config.Driver, config.Open)
	return err
}

// InitTestDb initializes a database for testing purpose.
func InitTestDb(t *testing.T) {
	config := conf.GetDbConfig()

	err := InitDb(config)
	if err != nil {
		t.Fatal(err)
	}

	fixtures, err := ioutil.ReadFile("resources/fixtures/testdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	database.MustExec(string(fixtures))
}
