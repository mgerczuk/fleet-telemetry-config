package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
)

type Application struct {
	AppName      string  `json:"app_name,omitempty"`
	ClientId     string  `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty"`
}

type Keys struct {
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
}

type User struct {
	Name  string                `json:"name"`
	Token *tesla_api.FleetToken `json:"token,omitempty"`
}

type FleetTelemetryConfig struct {
	Vins   []string                           `json:"vins"`
	Config tesla_api.FleetTelemetryConfigData `json:"config"`
}

type Persist struct {
	Application          Application          `json:"application,omitempty"`
	Keys                 Keys                 `json:"keys,omitempty"`
	Users                map[string]*User     `json:"users,omitempty"`
	FleetTelemetryConfig FleetTelemetryConfig `json:"fleet_telemetry_config"`
}

var persistFile string

func initFilename() {
	if persistFile == "" {
		flag.StringVar(&persistFile, "persist", "persist.json", "application persistent data")
		flag.Parse()
	}
}

func GetPersist() (data *Persist, err error) {
	initFilename()

	file, err := os.Open(persistFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, err
}

func PutPersist(data *Persist) error {
	initFilename()

	jsonString, _ := json.Marshal(data)
	return os.WriteFile(persistFile, jsonString, os.ModePerm)
}
