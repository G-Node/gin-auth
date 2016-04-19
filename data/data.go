// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package data

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // pg driver needs to be imported in order to load it
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
)

var database *sqlx.DB

// DbConf contains data needed to connect to a SQL database.
// The struct contains yaml annotations in order to be compatible with gooses
// database configuration file (conf/dbconf.yml)
type DbConf struct {
	Driver string `yaml:"driver"`
	Open   string `yaml:"open"`
}

// InitDb initializes a global database connection.
// An existing connection will be closed.
func InitDb(conf *DbConf) (err error) {
	if database != nil {
		database.Close()
	}
	database, err = sqlx.Connect(conf.Driver, conf.Open)
	return err
}

// LoadDbConf loads a database configuration from a yaml file.
func LoadDbConf(path string) (*DbConf, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &DbConf{}
	err = yaml.Unmarshal(content, conf)

	return conf, err
}

// InitTestDb initializes a database for testing purpose.
func InitTestDb(t *testing.T) {
	conf, err := LoadDbConf("../conf/dbconf.yml")
	if err != nil {
		t.Fatal(err)
	}

	err = InitDb(conf)
	if err != nil {
		t.Fatal(err)
	}

	fixtures, err := ioutil.ReadFile("../conf/fixtures/testdb.sql")
	if err != nil {
		t.Fatal(err)
	}

	database.MustExec(string(fixtures))
}
