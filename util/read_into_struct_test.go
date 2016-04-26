// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"testing"
)

type testData struct {
	SomeString  string
	AnInteger   int
	Unsigned    uint
	StringSlice []string
	Boolean     bool
	AndAFloat   float32
}

func TestReadMapIntoStruct(t *testing.T) {
	source := map[string][]string{
		"some_string":  []string{"foo"},
		"an_integer":   []string{"404"},
		"unsigned":     []string{"111"},
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
