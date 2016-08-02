// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package conf

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/smtp"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

// The unit of all life times and intervals is minute
const (
	defaultSessionLifeTime       = 2880
	defaultTokenLifeTime         = 1440
	defaultGrantReqLifeTime      = 15
	defaultUnusedAccountLifeTime = 10080
	defaultCleanerInterval       = 15
)

// Default smtp settings
const (
	defaultPort = 587
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
	if basePath != "" {
		resourcesPath = path.Join(basePath, "src", "github.com", "G-Node", "gin-auth", "resources")
	}
}

// SetResourcesPath sets the resource path to the specified location.
// This function should only be used before other helpers from the conf
// package are used.
func SetResourcesPath(res string) {
	resourcesPath = res
}

// ServerConfig provides several general configuration parameters for gin-auth
type ServerConfig struct {
	Host                  string
	Port                  int
	BaseURL               string
	SessionLifeTime       time.Duration
	TokenLifeTime         time.Duration
	GrantReqLifeTime      time.Duration
	UnusedAccountLifeTime time.Duration
	CleanerInterval       time.Duration
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

// SmtpCredentials contains the credentials required to send e-mails
// via smtp. Mode constitutes a switch whether e-mails should actually be sent or not.
// Supported values of Mode are: print and skip; print will write the content of
// any e-mail to the commandline / log, skip will skip over any e-mail sending process.
// For any other value of "Mode" e-mails will be sent.
type SmtpCredentials struct {
	From     string
	Username string
	Password string
	Host     string
	Port     int
	Mode     string
}

var smtpCred *SmtpCredentials
var smtpCredLock = sync.Mutex{}

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
			Http struct {
				Host                  string `yaml:"Host"`
				Port                  int    `yaml:"Port"`
				BaseURL               string `yaml:"BaseURL"`
				SessionLifeTime       int    `yaml:"SessionLifeTime"`
				TokenLifeTime         int    `yaml:"TokenLifeTime"`
				GrantReqLifeTime      int    `yaml:"GrantReqLifeTime"`
				UnusedAccountLifeTime int    `yaml:"UnusedAccountLifeTime"`
				CleanerInterval       int    `yaml:"CleanerInterval"`
			}
		}{}
		err = yaml.Unmarshal(content, config)
		if err != nil {
			panic(err)
		}

		// set defaults
		if config.Http.BaseURL == "" {
			if config.Http.Port == 80 {
				config.Http.BaseURL = fmt.Sprintf("http://%s", config.Http.Host)
			} else {
				config.Http.BaseURL = fmt.Sprintf("http://%s:%d", config.Http.Host, config.Http.Port)
			}
		}
		if config.Http.SessionLifeTime == 0 {
			config.Http.SessionLifeTime = defaultSessionLifeTime
		}
		if config.Http.TokenLifeTime == 0 {
			config.Http.TokenLifeTime = defaultTokenLifeTime
		}
		if config.Http.GrantReqLifeTime == 0 {
			config.Http.GrantReqLifeTime = defaultGrantReqLifeTime
		}
		if config.Http.UnusedAccountLifeTime == 0 {
			config.Http.UnusedAccountLifeTime = defaultUnusedAccountLifeTime
		}
		if config.Http.CleanerInterval == 0 {
			config.Http.CleanerInterval = defaultCleanerInterval
		}

		serverConfig = &ServerConfig{
			Host:                  config.Http.Host,
			Port:                  config.Http.Port,
			BaseURL:               config.Http.BaseURL,
			SessionLifeTime:       time.Duration(config.Http.SessionLifeTime) * time.Minute,
			TokenLifeTime:         time.Duration(config.Http.TokenLifeTime) * time.Minute,
			GrantReqLifeTime:      time.Duration(config.Http.GrantReqLifeTime) * time.Minute,
			UnusedAccountLifeTime: time.Duration(config.Http.UnusedAccountLifeTime) * time.Minute,
			CleanerInterval:       time.Duration(config.Http.CleanerInterval) * time.Minute,
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

// GetResourceFile returns the path to a resource file using the global resource path.
// The path will be constructed from the resource path and all given path elements in p.
func GetResourceFile(p ...string) string {
	tmp := make([]string, 1, len(p)+1)
	tmp[0] = resourcesPath
	tmp = append(tmp, p...)
	return path.Join(tmp...)
}

// GetClientsConfigFile returns the path to the clients configuration file.
func GetClientsConfigFile() string {
	return path.Join(resourcesPath, clientsConfigFile)
}

// GetStaticFilesDir returns the path to the static files directory.
func GetStaticFilesDir() string {
	return path.Join(resourcesPath, staticFilesDir)
}

// GetSmtpCredentials loads the smtp access information from a yaml file when called the first time.
// Returns a struct with the smtp credentials.
func GetSmtpCredentials() *SmtpCredentials {
	smtpCredLock.Lock()
	defer smtpCredLock.Unlock()

	if smtpCred == nil {
		content, err := ioutil.ReadFile(path.Join(resourcesPath, serverConfigFile))
		if err != nil {
			panic(err)
		}

		credentials := &struct {
			Smtp struct {
				From     string `yaml:"From"`
				Username string `yaml:"Username"`
				Password string `yaml:"Password"`
				Host     string `yaml:"Host"`
				Port     int    `yaml:"Port"`
				Mode     string `yaml:"Mode"`
			}
		}{}
		err = yaml.Unmarshal(content, credentials)
		if err != nil {
			panic(err)
		}

		if credentials.Smtp.Port == 0 {
			credentials.Smtp.Port = defaultPort
		}

		smtpCred = &SmtpCredentials{
			From:     credentials.Smtp.From,
			Username: credentials.Smtp.Username,
			Password: credentials.Smtp.Password,
			Host:     credentials.Smtp.Host,
			Port:     credentials.Smtp.Port,
			Mode:     credentials.Smtp.Mode,
		}
	}

	return smtpCred
}

// SmtpCheck tests whether a connection to the specified smtp server can be established
// with the provided credentials and will panic if it cannot.
func SmtpCheck() {
	cred := GetSmtpCredentials()
	if cred.Mode == "skip" || cred.Mode == "print" {
		return
	}

	addr := cred.Host + ":" + strconv.Itoa(cred.Port)
	auth := smtp.PlainAuth("", cred.Username, cred.Password, cred.Host)

	netCon, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		panic(err.Error())
	}
	if err = netCon.Close(); err != nil {
		panic(err.Error())
	}

	c, err := smtp.Dial(addr)
	if err != nil {
		panic(err.Error())
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: cred.Host}
		if err = c.StartTLS(config); err != nil {
			panic(err.Error())
		}
	}

	if err = c.Auth(auth); err != nil {
		panic(err.Error())
	}

	if err = c.Quit(); err != nil {
		panic(err.Error())
	}
}
