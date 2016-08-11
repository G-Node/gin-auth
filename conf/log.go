// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package conf

import (
	"io"
	"os"

	"github.com/NYTimes/logrotate"
	"github.com/Sirupsen/logrus"
)

var logEnv *LogEnv

// Logging environment with error and access log and a function to
// defer closing any associated files.
type LogEnv struct {
	Err    *logrus.Logger
	Access *logrus.Logger
	Close  func()
}

// InitLogEnv initializes loggers for access and error.
// Default access log directs to Stdout, default error log
// directs to Stderr. If log files are provided, the output
// will be directed to the respective default and the log file.
// Log files are opened using a logrotate compatible library.
func InitLogEnv() {

	accFile := "gin-auth.access.log"
	errFile := "gin-auth.err.log"

	logEnv = &LogEnv{
		Access: logrus.New(),
		Err:    logrus.New(),
	}
	logEnv.Access.Out = os.Stdout

	fs := make([]*logrotate.File, 0, 2)

	if accFile != "" {
		af, err := logrotate.NewFile(accFile)
		if err != nil {
			panic(err)
		}
		logEnv.Access.Out = io.MultiWriter(os.Stdout, af)
		fs = append(fs, af)
	}

	if errFile != "" {
		ef, err := logrotate.NewFile(errFile)
		if err != nil {
			panic(err)
		}
		logEnv.Err.Out = io.MultiWriter(os.Stderr, ef)
		fs = append(fs, ef)
	}

	logEnv.Close = func() {
		for _, f := range fs {
			f.Close()
		}
	}

	logEnv.Access.Info("Access logging started")
	logEnv.Err.Error("Error logging started")
}

