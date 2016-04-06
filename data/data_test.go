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
)

func failOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal(r)
	}
}

func initTestDb(t *testing.T) {
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
