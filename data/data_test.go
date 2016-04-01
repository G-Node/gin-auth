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
