package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mgerczuk/fleet-telemetry-config/config"
	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
)

func SendTelemetryConfig(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data, err := config.GetPersist()
		if err != nil {
			http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
			return
		}

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
		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, exists := data.Users[params.Uid]
		if !exists {
			http.Error(w, fmt.Sprintf("uid '%s' not found", params.Uid), http.StatusNotFound)
			return
		}

		if user.Token == nil {
			http.Error(w, "no token available. Create at /auth/request", http.StatusInternalServerError)
			return
		}

		var telemetryConfig tesla_api.FleetTelemetryConfigData

		telemetryConfig.Port = params.Config.Port
		telemetryConfig.Exp = params.Config.Exp
		telemetryConfig.AlertTypes = params.Config.AlertTypes
		telemetryConfig.Fields = params.Config.Fields

		telemetryConfig.Hostname = configData.PublicHostname

		buf, err := os.ReadFile(configData.PublicCert)
		if err != nil {
			http.Error(w, "cannot access cert", http.StatusInternalServerError)
			return
		}
		telemetryConfig.Ca = string(buf)

		client := tesla_api.NewVehicleClient(configData.Audience, user.Token.AccessToken)
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

		data, err := config.GetPersist()
		if err != nil {
			http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
			return
		}

		uid := r.URL.Query().Get("uid")
		if uid == "" {
			http.Error(w, "missing parameter 'uid'", http.StatusBadRequest)
			return
		}

		user, exists := data.Users[uid]
		if !exists {
			http.Error(w, fmt.Sprintf("uid '%s' not found", uid), http.StatusNotFound)
			return
		}

		if user.Token == nil {
			http.Error(w, "no token available. Create at /auth/request", http.StatusInternalServerError)
			return
		}

		vin := r.URL.Query().Get("vin")
		if vin == "" {
			http.Error(w, "missing parameter 'vin'", http.StatusBadRequest)
			return
		}

		client := tesla_api.NewVehicleClient(configData.Audience, user.Token.AccessToken)

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
