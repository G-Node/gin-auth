// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// ValidationError provides information about each field of a struct that failed.
type ValidationError struct {
	Message     string
	FieldErrors map[string]string
}

// Error implements Go error
func (err *ValidationError) Error() string {
	return err.Message
}

// ReadQueryIntoStruct reads request query parameters into a struct with matching fields.
// Single query values are interpreted as coma separated list of values.
//
// See ReadMapIntoStruct for more information.
func ReadQueryIntoStruct(request *http.Request, dest interface{}, ignoreMissing bool) error {
	query := request.URL.Query()
	if query == nil {
		return errors.New("Request has no query parameters")
	}
	for k, v := range query {
		if len(v) == 1 {
			query[k] = strings.Split(v[0], ",")
		}
	}
	return ReadMapIntoStruct(query, dest, ignoreMissing)
}

// ReadFormIntoStruct call ParseForm on the request and reads all form data into a struct with
// matching fields.
//
// See ReadMapIntoStruct for more information.
func ReadFormIntoStruct(request *http.Request, dest interface{}, ignoreMissing bool) error {
	request.ParseForm()
	form := request.PostForm
	if form == nil {
		return errors.New("Request has no form data")
	}
	return ReadMapIntoStruct(form, dest, ignoreMissing)
}

// ReadMapIntoStruct reads values from a map of string slices into a struct with matching fields
// and converts/parses it as needed. ReadMapIntoStruct tries to find matching fields using the
// original field name or the original name converted into its snake case equivalent.
//
// Supported field types for the target struct are integers, floats, boolean and slices of strings.
// But the reader does not check whether a string representation of an int, or float actually fits
// into the destination field.
//
// On errors the reader returns a ValidationError.
func ReadMapIntoStruct(src map[string][]string, dest interface{}, ignoreMissing bool) error {
	destVal := reflect.ValueOf(dest).Elem()
	destTyp := destVal.Type()

	valErr := &ValidationError{Message: "Unable to set one or more fields", FieldErrors: make(map[string]string)}
	count := destTyp.NumField()
	for i := 0; i < count; i++ {
		fieldTyp := destTyp.Field(i)
		fieldVal := destVal.Field(i)
		fieldName := fieldTyp.Name
		fieldNameSnake := ToSnakeCase(fieldTyp.Name)

		if !fieldVal.CanSet() {
			valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s can not be set", fieldName)
			continue
		}

		srcVal, ok := src[fieldName]
		if !ok {
			srcVal, ok = src[fieldNameSnake]
		}
		if !ok {
			if !ignoreMissing {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
			}
			continue
		}

		t := fieldVal.Interface()
		switch t := t.(type) {
		default:
			panic(fmt.Sprintf("Destination fiels %s has an unsupproted type: '%T'", fieldVal, t))
		case bool:
			if len(srcVal) < 1 {
				if !ignoreMissing {
					valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
				}
				continue
			}
			if len(srcVal) > 1 {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s contains too many values", fieldName)
				continue
			}
			boolVal, err := strconv.ParseBool(srcVal[0])
			if err != nil {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s requires a '%T' but the value was '%s'", fieldName, t, srcVal[0])
				continue
			}
			fieldVal.SetBool(boolVal)
		case int, int8, int16, int32, int64:
			if len(srcVal) < 1 {
				if !ignoreMissing {
					valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
				}
				continue
			}
			if len(srcVal) > 1 {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s contains too many values", fieldName)
				continue
			}
			intVal, err := strconv.ParseInt(srcVal[0], 10, 64)
			if err != nil {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s requires a '%T' but the value was '%s'", fieldName, t, srcVal[0])
				continue
			}
			fieldVal.SetInt(intVal)
		case uint, uint8, uint16, uint32, uint64:
			if len(srcVal) < 1 {
				if !ignoreMissing {
					valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
				}
				continue
			}
			if len(srcVal) > 1 {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s contains too many values", fieldName)
				continue
			}
			uintVal, err := strconv.ParseUint(srcVal[0], 10, 64)
			if err != nil {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s requires a '%T' but the value was '%s'", fieldName, t, srcVal[0])
				continue
			}
			fieldVal.SetUint(uintVal)
		case float32, float64:
			if len(srcVal) < 1 {
				if !ignoreMissing {
					valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
				}
				continue
			}
			if len(srcVal) > 1 {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s contains too many values", fieldName)
				continue
			}
			floatVal, err := strconv.ParseFloat(srcVal[0], 64)
			if err != nil {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s requires a '%T' but the value was '%s'", fieldName, t, srcVal[0])
				continue
			}
			fieldVal.SetFloat(floatVal)
		case string:
			if len(srcVal) < 1 {
				if !ignoreMissing {
					valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
				}
				continue
			}
			if len(srcVal) > 1 {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s contains too many values", fieldName)
				continue
			}
			fieldVal.SetString(srcVal[0])
		case []string:
			fieldVal.Set(reflect.ValueOf(srcVal))
		case sql.NullString:
			if len(srcVal) < 1 {
				if !ignoreMissing {
					valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s was missing", fieldName)
				}
				continue
			}
			if len(srcVal) > 1 {
				valErr.FieldErrors[fieldName] = fmt.Sprintf("Field %s contains too many values", fieldName)
				continue
			}
			if srcVal[0] != "" {
				fieldVal.Set(reflect.ValueOf(sql.NullString{String: srcVal[0], Valid: true}))
			}
		}
	}

	if len(valErr.FieldErrors) > 0 {
		return valErr
	}

	return nil
}
