package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
	"time"
)

type ServerConfig struct {
	Host             string
	Port             int
	BaseURL          string
	SessionLifeTime  time.Duration
	TokenLifeTime    time.Duration
	GrantReqLifeTime time.Duration
}

const (
	serverConfigFile        = "resources/conf/server.yml"
	defaultSessionLifeTime  = 2880
	defaultTokenLifeTime    = 1440
	defaultGrantReqLifeTime = 15
)

var serverConfig *ServerConfig = nil
var serverConfigLock = sync.Mutex{}

func GetServerConfig() *ServerConfig {
	serverConfigLock.Lock()
	defer serverConfigLock.Unlock()

	if serverConfig == nil {
		content, err := ioutil.ReadFile(serverConfigFile)
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
