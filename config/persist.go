package config

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
	log "github.com/sirupsen/logrus"
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

const refreshDays = 60

type User struct {
	Name  string                `json:"name"`
	Token *tesla_api.FleetToken `json:"token,omitempty"`

	timer *time.Timer
}

func (u *User) SetToken(t *tesla_api.FleetToken) {
	u.Token = t
	u.startRefreshTimer()
}

func (u *User) startRefreshTimer() {

	if u.Token == nil {
		return
	}

	expires := u.Token.CreatedAt.Add(time.Hour * (24 * refreshDays))

	if u.timer != nil {
		u.timer.Stop()
		u.timer = nil
	}

	if time.Now().After(expires) {
		u.doRefreshToken()
	} else {
		duration := time.Until(expires)
		log.Infof("Token refresh in %v", duration)

		u.timer = time.AfterFunc(duration, func() {
			u.doRefreshToken()
		})
	}
}

func (u *User) doRefreshToken() {

	log.Infof("Automatic token refresh before refresh token expires")

	// this relies on the fact that the persistent data is a singleton und u is
	// an element of data.Users!
	data := LockPersist()
	defer data.Unlock()

	t, err := tesla_api.RefreshToken(data.Application.ClientId, u.Token.RefreshToken)
	if err != nil {
		return
	}
	u.SetToken(t)

	PutPersist(data)
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

	for _, u := range singleton.Users {
		u.startRefreshTimer()
	}

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
