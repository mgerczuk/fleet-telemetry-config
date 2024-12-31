package config

import (
	"encoding/json"
	"flag"
	"os"
)

type Config struct {
	PublicHostname string `json:"public_hostname,omitempty"`

	// PublicPort is the port of the web server facing the internet
	PublicPort int `json:"public_port,omitempty"`

	// Certificate for the PublicHostname
	PublicCert string `json:"cert,omitempty"`

	// Private key for the PublicHostname
	PublicKey string `json:"key,omitempty"`

	// PublicPort is the port of the web server facing the local network
	PrivatePort int `json:"private_port,omitempty"`

	// The folder of the static web pages
	WebRoot string `json:"web_root,omitempty"`

	// the base url for Tesla API
	Audience string `json:"audience,omitempty"`
}

func LoadApplicationConfiguration() (config *Config, err error) {
	configFilePath := loadConfigFlags()

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

func loadConfigFlags() string {
	applicationConfig := ""
	flag.StringVar(&applicationConfig, "config", "config.json", "application configuration file")

	flag.Parse()
	return applicationConfig
}
