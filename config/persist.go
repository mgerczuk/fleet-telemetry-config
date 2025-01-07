package config

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
)

type Application struct {
	AppName      string  `json:"app_name,omitempty"`
	ClientId     string  `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty"`

	// the base url for Tesla API see https://developer.tesla.com/docs/fleet-api/getting-started/base-urls
	// TODO: Is needed for app registration and vehicle configuration. Better store with user?
	Audience string `json:"audience,omitempty"`
}

type Keys struct {
	PrivateKey string `json:"private_key,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
}

type User struct {
	Name string `json:"name"`
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

	filename string     `json:"-"`
	mtx      sync.Mutex `json:"-"`
}

var singleton *Persist = nil

func (p *Persist) Unlock() {
	p.mtx.Unlock()
}

func InitPersist(persistFile string) error {

	file, err := os.Open(persistFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create new instance
			singleton = &Persist{filename: persistFile}
			return nil
		}
		return err
	}
	defer file.Close()

	persistData := Persist{filename: persistFile}
	err = json.NewDecoder(file).Decode(&persistData)
	if err != nil {
		return err
	}

	singleton = &persistData
	return nil
}

func LockPersist() (data *Persist) {
	singleton.mtx.Lock()
	return singleton
}

func PutPersist(data *Persist) error {

	jsonString, _ := json.Marshal(data)
	return os.WriteFile(data.filename, jsonString, os.ModePerm)
}
