package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mgerczuk/fleet-telemetry-config/auth"
	"github.com/mgerczuk/fleet-telemetry-config/config"
	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
)

type registerParams struct {
	Scope string `json:"scope"`
}

func Register(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := config.GetPersist()
		if err != nil {
			http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
			return
		}

		var params registerParams
		err = json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("data = %v", data)
		fleetToken, err := tesla_api.GetClientCredentials(data.Application.ClientId, *data.Application.ClientSecret, configData.Audience, params.Scope)
		fmt.Printf("cred = %v", fleetToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t2, err := tesla_api.Register(configData.Audience, fleetToken.AccessToken, configData.PublicHostname)
		fmt.Println(err)
		fmt.Println(t2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func GetInitialToken(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := config.GetPersist()
		if err != nil {
			http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
			return
		}

		var params struct {
			Uid  string `json:"uid"`
			Code string `json:"code"`
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

		fleetToken, err := tesla_api.GetAuthorizationCode(data.Application.ClientId, *data.Application.ClientSecret, configData.Audience, params.Code, auth.GetRedirectUri(configData.PublicHostname))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.Token = fleetToken
		config.PutPersist(data)

		json.NewEncoder(w).Encode(fleetToken)
	}
}

func RefreshToken(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data, err := config.GetPersist()
		if err != nil {
			http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
			return
		}

		var params struct {
			Uid string `json:"uid"`
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

		fleetToken, err := tesla_api.RefreshToken(data.Application.ClientId, user.Token.RefreshToken)
		if err != nil {
			http.Error(w, "cannot access persistent data", http.StatusInternalServerError)
			return
		}

		user.Token = fleetToken
		config.PutPersist(data)

		json.NewEncoder(w).Encode(fleetToken)
	}
}
