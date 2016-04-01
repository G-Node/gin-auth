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

func initTestDb(t *testing.T) {
	conf, err := LoadDbConf("../conf/dbconf.yml")
	if err != nil {
		t.Error(err)
	}

	err = InitDb(conf)
	if err != nil {
		t.Error(err)
	}

	fixtures, err := ioutil.ReadFile("../conf/fixtures/testdb.sql")
	if err != nil {
		t.Error(err)
	}

	database.MustExec(string(fixtures))
}
