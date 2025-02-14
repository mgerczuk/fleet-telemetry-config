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

		// Save data in local variables and unlock peristent data because
		// tesla_api.Register tries to GET the public key which will cause a
		// deadlock otherwise
		data := config.LockPersist()
		clientId := data.Application.ClientId
		clientSecret := *data.Application.ClientSecret
		audience := data.Application.Audience
		data.Unlock()

		var params registerParams
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fleetToken, err := tesla_api.GetClientCredentials(clientId, clientSecret, audience, params.Scope)
		fmt.Printf("cred = %v", fleetToken)
		if err != nil {
			fmt.Printf("GetClientCredentials failed: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response, err := tesla_api.Register(audience, fleetToken.AccessToken, configData.PublicServer.Hostname)
		if err != nil {
			fmt.Printf("Register failed: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(response)
	}
}

func GetInitialToken(configData config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data := config.LockPersist()
		defer data.Unlock()

		var params struct {
			Uid  string `json:"uid"`
			Code string `json:"code"`
		}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, exists := data.Users[params.Uid]
		if !exists {
			http.Error(w, fmt.Sprintf("uid '%s' not found", params.Uid), http.StatusNotFound)
			return
		}

		fleetToken, err := tesla_api.GetAuthorizationCode(data.Application.ClientId, *data.Application.ClientSecret, data.Application.Audience, params.Code, auth.GetRedirectUri(configData.PublicServer.Hostname))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user.SetToken(fleetToken)
		config.PutPersist(data)

		json.NewEncoder(w).Encode(fleetToken)
	}
}
