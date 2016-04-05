package util

import (
	"testing"
)

func TestRandomToken(t *testing.T) {
	token := RandomToken()
	if len(token) != 104 {
		t.Error("Token length is expected to be 104")
	}
}
