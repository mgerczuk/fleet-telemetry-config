package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	PublicServer struct {
		Hostname string `json:"hostname,omitempty"`

		// PublicPort is the port of the web server facing the internet
		Port int `json:"port,omitempty"`

		// Certificate for the PublicHostname
		Cert string `json:"cert,omitempty"`

		// Private key for the PublicHostname
		Key string `json:"key,omitempty"`
	} `json:"public_server"`

	PrivateServer struct {
		// PublicPort is the port of the web server facing the local network
		Port int `json:"port,omitempty"`

		// The folder of the static web pages
		WebRoot string `json:"web_root,omitempty"`
	} `json:"private_server"`
}

func LoadApplicationConfiguration(configFilePath string) (config *Config, err error) {

	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, err
	}

	return config, err
}
