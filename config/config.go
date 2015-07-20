// Package config reads configuration information from a file.
package config

import (
	"encoding/json"
	"io/ioutil"
)

const cfgFileName = "./zendesk-livestats.json"

type ZendeskConfig struct {
	BaseUrl  string
	Username string
	Password string
}

var cfg *ZendeskConfig = nil

// readConfig reads the config from `cfgFileName` and Unmarshals the data
// into a Config struct
func readConfig() *ZendeskConfig {
	cfg = &ZendeskConfig{}
	bytes, err := ioutil.ReadFile(cfgFileName)
	if err != nil {
		log.Error("Error while reading file %v: %+v", cfgFileName, err)
		return &ZendeskConfig{}
	}
	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		log.Error("Error unmarshaling config file json: %+v", err)
		panic(err)
	}
	return cfg
}

// GetConfig determines whether the config singleton is already initialized.
// If it is initialized, it is simply returned. Else, a call to readConfig is made.
func GetConfig() ZendeskConfig {
	if cfg == nil {
		cfg = readConfig()
	}
	return *cfg
}
