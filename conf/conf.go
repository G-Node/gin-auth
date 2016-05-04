// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

const (
	defaultSessionLifeTime  = 2880
	defaultTokenLifeTime    = 1440
	defaultGrantReqLifeTime = 15
)

var (
	resourcesPath     string
	serverConfigFile  = path.Join("conf", "server.yml")
	dbConfigFile      = path.Join("conf", "dbconf.yml")
	clientsConfigFile = path.Join("conf", "clients.yml")
	staticFilesDir    = path.Join("static")
)

func init() {
	basePath := os.Getenv("GOPATH")
	if basePath == "" {
		basePath = "./"
	}
	resourcesPath = path.Join(basePath, "src", "github.com", "G-Node", "gin-auth", "resources")
}

// SetResourcesPath sets the resource path to the specified location.
// This function should only be used before other helpers from the conf
// package are used.
func SetResourcesPath(res string) {
	resourcesPath = res
}

// ServerConfig provides several general configuration parameters for gin-auth
type ServerConfig struct {
	Host             string
	Port             int
	BaseURL          string
	SessionLifeTime  time.Duration
	TokenLifeTime    time.Duration
	GrantReqLifeTime time.Duration
}

var serverConfig *ServerConfig
var serverConfigLock = sync.Mutex{}

// DbConfig contains data needed to connect to a SQL database.
// The struct contains yaml annotations in order to be compatible with gooses
// database configuration file (resources/conf/dbconf.yml)
type DbConfig struct {
	Driver string `yaml:"driver"`
	Open   string `yaml:"open"`
}

var dbConfig *DbConfig
var dbConfigLock = sync.Mutex{}

// GetServerConfig loads the server configuration from a yaml file when called the first time.
// Returns a struct with configuration information.
func GetServerConfig() *ServerConfig {
	serverConfigLock.Lock()
	defer serverConfigLock.Unlock()

	if serverConfig == nil {
		content, err := ioutil.ReadFile(path.Join(resourcesPath, serverConfigFile))
		if err != nil {
			panic(err)
		}

		config := &struct {
			Host             string `yaml:"Host"`
			Port             int    `yaml:"Port"`
			BaseURL          string `yaml:"BaseURL"`
			SessionLifeTime  int    `yaml:"SessionLifeTime"`
			TokenLifeTime    int    `yaml:"TokenLifeTime"`
			GrantReqLifeTime int    `yaml:"GrantReqLifeTime"`
		}{}
		err = yaml.Unmarshal(content, config)
		if err != nil {
			panic(err)
		}

		// set defaults
		if config.BaseURL == "" {
			if config.Port == 80 {
				config.BaseURL = fmt.Sprintf("http://%s", config.Host)
			} else {
				config.BaseURL = fmt.Sprintf("http://%s:%d", config.Host, config.Port)
			}
		}
		if config.SessionLifeTime == 0 {
			config.SessionLifeTime = defaultSessionLifeTime
		}
		if config.TokenLifeTime == 0 {
			config.TokenLifeTime = defaultTokenLifeTime
		}
		if config.GrantReqLifeTime == 0 {
			config.GrantReqLifeTime = defaultGrantReqLifeTime
		}

		serverConfig = &ServerConfig{
			Host:             config.Host,
			Port:             config.Port,
			BaseURL:          config.BaseURL,
			SessionLifeTime:  time.Duration(config.SessionLifeTime) * time.Minute,
			TokenLifeTime:    time.Duration(config.TokenLifeTime) * time.Minute,
			GrantReqLifeTime: time.Duration(config.TokenLifeTime) * time.Minute,
		}
	}

	return serverConfig
}

// GetDbConfig loads a database configuration from a yaml file when called the first time.
// Returns a struct with configuration information.
func GetDbConfig() *DbConfig {
	dbConfigLock.Lock()
	defer dbConfigLock.Unlock()

	if dbConfig == nil {
		content, err := ioutil.ReadFile(path.Join(resourcesPath, dbConfigFile))
		if err != nil {
			panic(err)
		}

		config := &DbConfig{}
		err = yaml.Unmarshal(content, config)
		if err != nil {
			panic(err)
		}

		dbConfig = config
	}

	return dbConfig
}

func GetResourceFile(p ...string) string {
	tmp := make([]string, 1, len(p)+1)
	tmp[0] = resourcesPath
	tmp = append(tmp, p...)
	return path.Join(tmp...)
}

func GetClientsConfigFile() string {
	return path.Join(resourcesPath, clientsConfigFile)
}

func GetStaticFilesDir() string {
	return path.Join(resourcesPath, staticFilesDir)
}
