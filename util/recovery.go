// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/Sirupsen/logrus"
)

type recoveryHandler struct {
	handler    http.Handler
	logger     interface{}
	printStack bool
}

// RecoveryHandler recovers from a panic, writes an HTTP InternalServerError,
// logs the panic message to the defined logging mechanism and continues
// to the next handler.
func RecoveryHandler(h http.Handler, l interface{}, ps bool) http.Handler {
	return &recoveryHandler{
		handler:    h,
		logger:     l,
		printStack: ps,
	}
}

func (h recoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log(err)
		}
	}()

	h.handler.ServeHTTP(w, req)
}

func (h recoveryHandler) log(msg interface{}) {
	if h.logger != nil {
		switch h.logger.(type) {
		default:
			l := log.New(os.Stderr, "", log.LstdFlags)
			l.Println(msg)
		case *logrus.Logger:
			h.logger.(*logrus.Logger).Error(msg)
		}
	}

	if h.printStack {
		debug.PrintStack()
	}
}
