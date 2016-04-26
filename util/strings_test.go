// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"sort"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	var str string
	str = ToSnakeCase("Foo")
	if str != "foo" {
		t.Error("String expected to be 'foo'")
	}
	str = ToSnakeCase("FooBar")
	if str != "foo_bar" {
		t.Error("String expected to be 'foo_bar'")
	}
	str = ToSnakeCase("Bond007")
	if str != "bond007" {
		t.Error("String expected to be 'bond007'")
	}
	str = ToSnakeCase("MyUUID")
	if str != "my_uuid" {
		t.Error("String expected to be 'my_uuid'")
	}
	str = ToSnakeCase("i_hate_camels")
	if str != "i_hate_camels" {
		t.Error("String expected to be 'i_hate_camels'")
	}
}

func TestStringSet(t *testing.T) {
	set := NewStringSet("a", "b")
	if !set.Contains("a") {
		t.Error("Set should contain 'a'")
	}
	if !set.Contains("b") {
		t.Error("Set should contain 'b'")
	}
	if set.Contains("c") {
		t.Error("Set should not contain 'c'")
	}
	if set.Len() != 2 {
		t.Error("Set length expected to be 2")
	}

	set = set.Add("c")
	if !set.Contains("a") {
		t.Error("Set should contain 'a'")
	}
	if !set.Contains("b") {
		t.Error("Set should contain 'b'")
	}
	if !set.Contains("c") {
		t.Error("Set should contain 'c'")
	}
	if set.Len() != 3 {
		t.Error("Set length expected to be 2")
	}

	set = set.Add("a")
	if set.Len() != 3 {
		t.Error("Set length expected to be 2")
	}
}

func TestStringSetIsSuperset(t *testing.T) {
	super := NewStringSet("apple", "banana", "strawberry")
	sub := NewStringSet("apple", "strawberry")
	if !super.IsSuperset(sub) {
		t.Error("Should be a superset")
	}
	if sub.IsSuperset(super) {
		t.Error("Should not be a superset")
	}
}

func TestStringSetUnion(t *testing.T) {
	set1 := NewStringSet("apple", "banana", "strawberry")
	set2 := NewStringSet("apple", "blueberry", "strawberry")
	uni := set1.Union(set2)
	for _, s := range []string{"apple", "banana", "blueberry", "strawberry"} {
		if !uni.Contains(s) {
			t.Errorf("Union should contain '%s'", s)
		}
	}
}

func TestStringSetStrings(t *testing.T) {
	set := NewStringSet("bar", "foo", "bla")
	sorted := sort.StringSlice(set.Strings())
	sorted.Sort()
	for _, s := range []string{"bar", "bla", "foo"} {
		if sorted.Search(s) >= 3 {
			t.Errorf("'%s' was not found", s)
		}
	}
}

func TestStringSetScan(t *testing.T) {
	set := NewStringSet()
	set.Scan([]byte(`{"foo","\"bar",bla,"\\blub"}`))

	if set.Len() != 4 {
		t.Error("The set should contain four elements")
	}
	if !set.Contains(`foo`) {
		t.Errorf(`Set does not contain 'foo'`)
	}
	if !set.Contains(`\"bar`) {
		t.Errorf(`Set does not contain '\"bar'`)
	}
	if !set.Contains(`bla`) {
		t.Errorf(`Set does not contain 'bla'`)
	}
	if !set.Contains(`foo`) {
		t.Errorf(`Set does not contain 'foo'`)
	}
	if !set.Contains(`\\blub`) {
		t.Errorf(`Set does not contain '\\blub'`)
	}
}

func TestStringSetValue(t *testing.T) {
	slice := NewStringSet(`bar\`, `blub"`, `foo`)

	value, err := slice.Value()
	if err != nil {
		t.Error(err)
	}
	str, ok := value.(string)
	if !ok {
		t.Error("Unable to converto into bytes")
	}
	if str != `{"bar\\","blub\"","foo"}` {
		t.Error(`str was supposed to be '{"bar\\","blub\"","foo"}'`)
	}
}
