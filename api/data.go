package api

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mgerczuk/fleet-telemetry-config/config"
)

func HandleDataModel(muxPrivate *http.ServeMux, configData *config.Config) {
	muxPrivate.HandleFunc("GET /api/data/config", getConfig(*configData))
	muxPrivate.HandleFunc("/api/data/application", getApplication)
	muxPrivate.HandleFunc("/api/data/keys", getKeys)
	muxPrivate.HandleFunc("/api/data/users", handleUsers)
	muxPrivate.HandleFunc("GET /api/data/token_expires", getTokenExpires)
	muxPrivate.HandleFunc("/api/data/telemetry_config", getTelemetryConfig)
	muxPrivate.HandleFunc("POST /api/data/challenge", postChallenge)
}

func getConfig(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(configData)
	}
}

func getApplication(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

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
		data.Application.Audience = app.Audience
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

	data := config.LockPersist()
	defer data.Unlock()

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

	case "POST":
		var err error
		data.Keys.PrivateKey, data.Keys.PublicKey, err = createKeys()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = config.PutPersist(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func createKeys() (privateKeyPEM string, publicKeyPEM string, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	privatePemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	var privateKeyRow bytes.Buffer
	err = pem.Encode(&privateKeyRow, privatePemBlock)
	if err != nil {
		return "", "", err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	publicPemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	var publicKeyRow bytes.Buffer
	err = pem.Encode(&publicKeyRow, publicPemBlock)
	if err != nil {
		return "", "", err
	}

	return privateKeyRow.String(), publicKeyRow.String(), nil
}

func handleUsers(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

	switch method := r.Method; method {
	case "GET":
		type user struct {
			Uid  string `json:"uid"`
			Name string `json:"name"`
		}

		names := make([]user, 0, len(data.Users))
		for key, u := range data.Users {
			names = append(names, user{Uid: key, Name: u.Name})
		}

		json.NewEncoder(w).Encode(names)

	case "POST":
		type user struct {
			Name string `json:"name"`
		}
		var u user
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		id := uuid.New()
		if data.Users == nil {
			data.Users = make(map[string]*config.User)
		}
		data.Users[id.String()] = &config.User{Name: u.Name}

		err = config.PutPersist(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func getTokenExpires(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

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

	tm, _ := user.Token.ExpirationTime()

	result := map[string]interface{}{
		"expires_at": tm.String(),
	}

	json.NewEncoder(w).Encode(result)
}

func getTelemetryConfig(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

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

var lastChallenge string

func postChallenge(w http.ResponseWriter, r *http.Request) {

	var param struct {
		Challenge string `json:"challenge"`
	}
	err := json.NewDecoder(r.Body).Decode(&param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lastChallenge = param.Challenge
}

func GetChallenge(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Access-Control-Allow-Origin", "*")
	response := map[string]string{
		"challenge": lastChallenge,
	}
	json.NewEncoder(w).Encode(response)
}

func GetPublicKey(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

	w.Write([]byte(data.Keys.PublicKey))
}
