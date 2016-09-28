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
	"path/filepath"
	"strconv"
	"strings"
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
	defaultMailQueueInterval     = 1
)

// Default smtp settings
const (
	defaultPort = 587
)

var (
	resourcesPath     string
	configPath        string
	serverConfigFile  = "server.yml"
	dbConfigFile      = "dbconf.yml"
	clientsConfigFile = "clients.yml"
)

func init() {
	basePath := os.Getenv("GOPATH")
	if basePath != "" {
		resourcesPath = filepath.Join(basePath, "src", "github.com", "G-Node", "gin-auth", "resources")
		configPath = filepath.Join(resourcesPath, "conf")
	}
}

// SetResourcesPath sets the resource path to the specified location.
// This function should only be used before other helpers from the conf
// package are used.
func SetResourcesPath(res string) {
	resourcesPath = res
}

// SetConfigPath sets the resource path to the specified location.
func SetConfigPath(config string) {
	configPath = config
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
	MailQueueInterval     time.Duration
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

// LogLocations contains paths to the Access and Error log files.
type LogLocations struct {
	Access string
	Error  string
}

var logLoc *LogLocations
var logLocLock = sync.Mutex{}

// GetServerConfig loads the server configuration from a yaml file when called the first time.
// Returns a struct with configuration information.
func GetServerConfig() *ServerConfig {
	serverConfigLock.Lock()
	defer serverConfigLock.Unlock()

	if serverConfig == nil {
		content, err := ioutil.ReadFile(filepath.Join(configPath, serverConfigFile))
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
				MailQueueInterval     int    `yaml:"MailQueueInterval"`
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
		if config.Http.MailQueueInterval == 0 {
			config.Http.MailQueueInterval = defaultMailQueueInterval
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
			MailQueueInterval:     time.Duration(config.Http.MailQueueInterval) * time.Minute,
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
		content, err := ioutil.ReadFile(filepath.Join(configPath, dbConfigFile))
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
	return filepath.Join(tmp...)
}

// GetClientsConfigFile returns the path to the clients configuration file.
func GetClientsConfigFile() string {
	return filepath.Join(configPath, clientsConfigFile)
}

// GetSmtpCredentials loads the smtp access information from a yaml file when called the first time.
// Returns a struct with the smtp credentials.
func GetSmtpCredentials() *SmtpCredentials {
	smtpCredLock.Lock()
	defer smtpCredLock.Unlock()

	if smtpCred == nil {
		content, err := ioutil.ReadFile(filepath.Join(configPath, serverConfigFile))
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

// NoAuth is a minimal implementation of the smtp.Auth interface.
type NoAuth struct{}

// Start always returns proto = "" to indicate, that the authentication can be skipped.
func (a *NoAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	return "", nil, nil
}

// Next always returns toServer = nil to indicate, that the authentication can be skipped.
func (a *NoAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	return nil, nil
}

// SmtpCheck tests whether a connection to the specified smtp server can be established
// with the provided credentials and will panic if it cannot.
func SmtpCheck() error {
	cred := GetSmtpCredentials()
	if strings.ToLower(cred.Mode) == "skip" || strings.ToLower(cred.Mode) == "print" {
		return nil
	}

	addr := cred.Host + ":" + strconv.Itoa(cred.Port)
	netCon, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		return err
	}
	defer netCon.Close()

	c, err := smtp.NewClient(netCon, cred.Host)
	if err != nil {
		return err
	}

	var auth smtp.Auth
	if cred.Username == "" && cred.Password == "" {
		auth = &NoAuth{}
	} else {
		auth = smtp.PlainAuth("", cred.Username, cred.Password, cred.Host)
		if ok, _ := c.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: cred.Host}
			if err = c.StartTLS(config); err != nil {
				return err
			}
		}
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	if err = c.Quit(); err != nil {
		return err
	}
	return nil
}

// GetLogLocation loads log file locations from a yaml file when called the first time.
// Returns a struct with the log file locations.
func GetLogLocation() *LogLocations {
	logLocLock.Lock()
	defer logLocLock.Unlock()

	if logLoc == nil {
		fc, err := ioutil.ReadFile(filepath.Join(configPath, serverConfigFile))
		if err != nil {
			panic(err)
		}

		cont := &struct {
			Log struct {
				Access string `yaml:"Access"`
				Error  string `yaml:"Error"`
			}
		}{}
		err = yaml.Unmarshal(fc, cont)
		if err != nil {
			panic(err)
		}
		logLoc = &LogLocations{
			Access: cont.Log.Access,
			Error:  cont.Log.Error,
		}
	}

	return logLoc
}
