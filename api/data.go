package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mgerczuk/fleet-telemetry-config/config"
)

func HandleDataModel(muxPrivate *http.ServeMux, configData *config.Config) {
	muxPrivate.HandleFunc("GET /api/data/config", getConfig(*configData))
	muxPrivate.HandleFunc("/api/data/application", getApplication)
	muxPrivate.HandleFunc("/api/data/keys", getKeys)
	muxPrivate.HandleFunc("GET /api/data/users", getUsers)
	muxPrivate.HandleFunc("GET /api/data/token_expires", getTokenExpires)
	muxPrivate.HandleFunc("/api/data/telemetry_config", getTelemetryConfig)
}

func getConfig(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(configData)
	}
}

func getApplication(w http.ResponseWriter, r *http.Request) {
	data, err := config.GetPersist()

	if err != nil {
		http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
		return
	}

	switch method := r.Method; method {
	case "GET":
		json.NewEncoder(w).Encode(data.Application)

	case "PUT":
		var app config.Application
		err := json.NewDecoder(r.Body).Decode(&app)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.Application.AppName = app.AppName
		data.Application.ClientId = app.ClientId
		data.Application.ClientSecret = app.ClientSecret
		err = config.PutPersist(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func getKeys(w http.ResponseWriter, r *http.Request) {
	data, err := config.GetPersist()

	if err != nil {
		http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
		return
	}

	switch method := r.Method; method {
	case "GET":
		json.NewEncoder(w).Encode(data.Keys)

	case "PUT":
		var keys config.Keys
		err := json.NewDecoder(r.Body).Decode(&keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.Keys.PrivateKey = keys.PrivateKey
		data.Keys.PublicKey = keys.PublicKey
		err = config.PutPersist(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {

	data, err := config.GetPersist()
	if err != nil {
		http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
		return
	}

	type user struct {
		Uid  string `json:"uid"`
		Name string `json:"name"`
	}

	names := make([]user, 0, len(data.Users))
	for key, u := range data.Users {
		names = append(names, user{Uid: key, Name: u.Name})
	}

	json.NewEncoder(w).Encode(names)
}

func getTokenExpires(w http.ResponseWriter, r *http.Request) {

	data, err := config.GetPersist()
	if err != nil {
		http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
		return
	}

	userId := r.URL.Query().Get("uid")
	if userId == "" {
		http.Error(w, "missing parameter 'uid'", http.StatusBadRequest)
		return
	}

	user, exists := data.Users[userId]
	if !exists {
		http.Error(w, fmt.Sprintf("'uid' %s not found", userId), http.StatusNotFound)
		return
	}

	if user.Token == nil {
		http.Error(w, "no token available. Create at /auth/request", http.StatusInternalServerError)
		return
	}

	tm := user.Token.CreatedAt
	tm = tm.Add(time.Duration(user.Token.ExpiresIn) * time.Second)

	result := map[string]interface{}{
		"expires_at": tm.String(),
	}

	json.NewEncoder(w).Encode(result)
}

func getTelemetryConfig(w http.ResponseWriter, r *http.Request) {

	data, err := config.GetPersist()
	if err != nil {
		http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
		return
	}

	switch method := r.Method; method {
	case "GET":
		json.NewEncoder(w).Encode(data.FleetTelemetryConfig)

	case "PUT":
		var keys config.FleetTelemetryConfig
		err := json.NewDecoder(r.Body).Decode(&keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		keys.Config.Ca = ""
		keys.Config.Hostname = ""

		data.FleetTelemetryConfig.Vins = keys.Vins
		data.FleetTelemetryConfig.Config = keys.Config
		err = config.PutPersist(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func GetPublicKey(w http.ResponseWriter, r *http.Request) {

	data, err := config.GetPersist()
	if err != nil {
		http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-Type", "application/x-pem-file")
	w.Write([]byte(data.Keys.PublicKey))
}
