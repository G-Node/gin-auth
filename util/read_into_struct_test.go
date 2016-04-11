package util

import (
	"testing"
)

type testData struct {
	SomeString  string
	AnInteger   int
	StringSlice []string
	Boolean     bool
	AndAFloat   float32
}

func TestReadMapIntoStruct(t *testing.T) {
	source := map[string][]string{
		"some_string":  []string{"foo"},
		"an_integer":   []string{"404"},
		"string_slice": []string{"foo", "bar"},
		"boolean":      []string{"true"},
		"AndAFloat":    []string{"3.1415"},
	}
	dest := &testData{}

	err := ReadMapIntoStruct(source, dest, false)
	if err != nil {
		t.Error(err)
	}

	delete(source, "some_string")
	err = ReadMapIntoStruct(source, dest, true)
	if err != nil {
		t.Error(err)
	}

	err = ReadMapIntoStruct(source, dest, false)
	if err == nil {
		t.Error("Error expected")
	} else {
		if err, ok := err.(*ValidationError); ok {
			_, ok := err.FieldErrors["SomeString"]
			if !ok {
				t.Error("Field error about 'SomeString' expected")
			}
		}
	}
}
