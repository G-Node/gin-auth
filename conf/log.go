// Copyright (c) 2016, German Neuroinformatics Node (G-Node)
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package conf

import (
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
