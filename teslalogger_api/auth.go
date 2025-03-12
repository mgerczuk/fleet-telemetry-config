package teslalogger_api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"

	"github.com/mgerczuk/fleet-telemetry-config/config"
	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
	log "github.com/sirupsen/logrus"
)

type ApiResult struct {
	Response         *string `json:"response"`
	Error            *string `json:"error"`
	ErrorDescription *string `json:"error_description"`
}

func ApiError(err error) ApiResult {
	msg := err.Error()
	return ApiResult{
		Response: nil,
		Error:    &msg,
	}
}

func ApiErrorString(err string) ApiResult {
	return ApiResult{
		Response: nil,
		Error:    &err,
	}
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

	bodyBytes, _ := io.ReadAll(r.Body)
	params, err := url.ParseQuery(string(bodyBytes))
	log.Infof("%v %v", params, err)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApiError(err))
		return
	}

	if !(params.Has("refresh_token") && params.Has("vin")) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApiErrorString("(400) parameter missing"))
		return
	}

	var user *config.User
	for _, u := range data.Users {
		if slices.Contains(u.Vins, params.Get("vin")) {
			user = u
			break
		}
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ApiErrorString("(404) vin not found"))
		return
	}

	t, err := tesla_api.RefreshToken(data.Application.ClientId, user.Token.RefreshToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	user.SetToken(t)
	config.PutPersist(data)

	json.NewEncoder(w).Encode(t)
}
