package data

import (
	"testing"
)

func TestSqlStringSliceScan(t *testing.T) {
	var slice SqlStringSlice
	slice.Scan([]byte(`{"foo","\"bar",bla,"\\blub"}`))
	if slice[0] != `foo` {
		t.Errorf(`slice[0] was supposed to be 'foo' but was '%s'`, slice[0])
	}
	if slice[1] != `\"bar` {
		t.Errorf(`slice[1] was supposed to be '\"bar' but was '%s'`, slice[1])
	}
	if slice[2] != `bla` {
		t.Errorf(`slice[2] was supposed to be 'bla' but was '%s'`, slice[2])
	}
	if slice[3] != `\\blub` {
		t.Errorf(`slice[3] was supposed to be '\\blub' but was '%s'`, slice[3])
	}
}

func TestSqlStringSliceValue(t *testing.T) {
	slice := SqlStringSlice{`foo`, `\bar`, `"blub`}
	value, err := slice.Value()
	if err != nil {
		t.Error(err)
	}
	str, ok := value.(string)
	if !ok {
		t.Error("Unable to converto into bytes")
	}
	if str != `{"foo","\\bar","\"blub"}` {
		t.Error(`str was supposed to be '{"foo","\\bar","\"blub"}'`)
	}
}
