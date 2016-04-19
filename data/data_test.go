package data

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Chdir("..")
	os.Exit(m.Run())
}
