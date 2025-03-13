package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mgerczuk/fleet-telemetry-config/config"
	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
)

func SendTelemetryConfig(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data := config.LockPersist()
		defer data.Unlock()

		var params struct {
			Uid    string   `json:"uid"`
			Vins   []string `json:"vins"`
			Config struct {
				Port       int                            `json:"port"`
				Exp        *int                           `json:"exp,omitempty"`
				AlertTypes []string                       `json:"alert_types"`
				Fields     map[string]tesla_api.FieldProp `json:"fields"`
			} `json:"config"`
		}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := GetValidAccessToken(data, params.Uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var telemetryConfig tesla_api.FleetTelemetryConfigData

		telemetryConfig.Port = params.Config.Port
		telemetryConfig.Exp = params.Config.Exp
		telemetryConfig.AlertTypes = params.Config.AlertTypes
		telemetryConfig.Fields = params.Config.Fields

		telemetryConfig.Hostname = configData.PublicServer.Hostname

		buf, err := os.ReadFile(configData.PublicServer.Cert)
		if err != nil {
			http.Error(w, "cannot access cert", http.StatusInternalServerError)
			return
		}
		telemetryConfig.Ca = string(buf)

		client := tesla_api.NewVehicleClient(data.Application.Audience, token)
		res, err := client.CreateFleetTelemetryConfig(&telemetryConfig, params.Vins, data.Keys.PrivateKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(res)
	}
}

func VehicleTelemetryConfig(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data := config.LockPersist()
		defer data.Unlock()

		uid := r.URL.Query().Get("uid")
		if uid == "" {
			http.Error(w, "missing parameter 'uid'", http.StatusBadRequest)
			return
		}

		token, err := GetValidAccessToken(data, uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		vin := r.URL.Query().Get("vin")
		if vin == "" {
			http.Error(w, "missing parameter 'vin'", http.StatusBadRequest)
			return
		}

		client := tesla_api.NewVehicleClient(data.Application.Audience, token)

		switch method := r.Method; method {
		case "GET":
			res, err := client.GetFleetTelemetryConfig(vin)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(res)

		case "DELETE":
			res, err := client.DeleteFleetTelemetryConfig(vin)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(res)

		default:
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		}
	}
}

func GetValidAccessToken(data *config.Persist, uid string) (string, error) {

	user, exists := data.Users[uid]
	if !exists {
		return "", errors.New(fmt.Sprintf("uid '%s' not found", uid))
	}

	if user.Token == nil {
		return "", errors.New("no token available. Create at /auth/request")
	}

	exp, err := user.Token.ExpirationTime()
	if err != nil {
		return "", err
	}
	expires := exp.Add(time.Minute * -10)

	if expires.Before(time.Now()) {
		_, t, err := tesla_api.RefreshToken(data.Application.ClientId, user.Token.RefreshToken)
		if err != nil {
			return "", err
		}
		user.SetToken(t)
		config.PutPersist(data)
	}

	return user.Token.AccessToken, nil
}
