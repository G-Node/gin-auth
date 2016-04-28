package conf

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Chdir("..")
	os.Exit(m.Run())
}

func TestGetServerConfig(t *testing.T) {
	config := GetServerConfig()
	if config.Host != "localhost" {
		t.Error("Host expected to be 'localhost'")
	}
	if config.Port != 8080 {
		t.Error("Port expected to be '8080'")
	}
	if config.BaseURL != "http://localhost:8080" {
		t.Error("BaseURL expected to be 'http://localhost:8080'")
	}
}
