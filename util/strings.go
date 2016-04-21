// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"database/sql/driver"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase turns a CamelCase string into its snake_case equivalent.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// StringInSlice returns true if a specific string was found in a slice or
// false otherwise.
func StringInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// StringSet provides some common set operations.
type StringSet interface {
	Add(string) StringSet
	Contains(string) bool
	IsSuperset(StringSet) bool
	Union(StringSet) StringSet
	Len() int
	Strings() []string
	Value() (driver.Value, error)
	Scan(interface{}) error
}

// NewStringSet creates a new StringSet from a slice of strings.
func NewStringSet(strs ...string) StringSet {
	set := mapStringSet{}
	for _, s := range strs {
		set[s] = true
	}
	return set
}

// Implements StringSet based on a sorted slice of strings.
type mapStringSet map[string]bool

// Add returns a set with one additional element.
func (set mapStringSet) Add(s string) StringSet {
	if !set.Contains(s) {
		fresh := mapStringSet{s: true}
		for s := range set {
			fresh[s] = true
		}
		return fresh
	}
	return set
}

// Contains returns true if a string is part of the set.
func (set mapStringSet) Contains(s string) bool {
	contains := set[s]
	return contains
}

// IsSuperset returns true if the set is a superset of the other set.
func (set mapStringSet) IsSuperset(other StringSet) bool {
	for _, s := range other.Strings() {
		if !set.Contains(s) {
			return false
		}
	}
	return true
}

// Union returns a set that contains all elements from both sets.
func (set mapStringSet) Union(other StringSet) StringSet {
	switch other.Len() {
	case 0:
		return set
	case 1:
		return set.Add(other.Strings()[0])
	default:
		fresh := set
		isCopy := false
		for _, s1 := range other.Strings() {
			if !set.Contains(s1) {
				if !isCopy {
					fresh := mapStringSet{}
					for s2 := range set {
						fresh[s2] = true
					}
				}
				fresh[s1] = true
			}
		}
		return fresh
	}
}

// Len returns the number of elements in the set.
func (set mapStringSet) Len() int {
	return len(set)
}

// Strings returns all elements as slice of strings.
func (set mapStringSet) Strings() []string {
	strs := make([]string, 0, len(set))
	for s := range set {
		strs = append(strs, s)
	}
	return strs
}

// Scan implements the Scanner interface.
func (set mapStringSet) Scan(src interface{}) error {
	var strs SqlStringSlice
	err := strs.Scan(src)
	if err != nil {
		return err
	}
	for s := range set {
		delete(set, s)
	}
	for _, s := range strs {
		set[s] = true
	}
	return nil
}

// Value implements the driver Valuer interface.
func (set mapStringSet) Value() (driver.Value, error) {
	return SqlStringSlice(set.Strings()).Value()
}
