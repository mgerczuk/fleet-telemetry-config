package teslalogger_api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"slices"

	"github.com/mgerczuk/fleet-telemetry-config/config"
	"github.com/mgerczuk/fleet-telemetry-config/tesla_api"
	log "github.com/sirupsen/logrus"
)

func asErrorObject(errmsg string) string {
	res := struct {
		Error string `json:"error"`
	}{
		Error: errmsg,
	}
	s, _ := json.Marshal(res)
	return string(s)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {

	data := config.LockPersist()
	defer data.Unlock()

	bodyBytes, _ := io.ReadAll(r.Body)
	params, err := url.ParseQuery(string(bodyBytes))
	log.Infof("/teslaredirect/refresh_token.php ParseQuery: params=%v error=%v", params, err)
	if err != nil {
		http.Error(w, asErrorObject(err.Error()), http.StatusBadRequest)
		return
	}

	if !params.Has("vin") {
		http.Error(w, asErrorObject("parameter 'vin' missing"), http.StatusBadRequest)
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
		http.Error(w, asErrorObject("vin not found"), http.StatusNotFound)
		return
	}

	statusCode, t, err := tesla_api.RefreshToken(data.Application.ClientId, user.Token.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	user.SetToken(t)
	config.PutPersist(data)

	json.NewEncoder(w).Encode(t)
}
