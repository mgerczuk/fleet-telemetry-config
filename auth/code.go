package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mgerczuk/fleet-telemetry-config/config"
)

const urlAuthorize = "https://auth.tesla.com/oauth2/v3/authorize"

type authRequest struct {
	state       string
	clientId    string
	redirectUri string
}

var currentRequest *authRequest

func StartAuth(config config.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		clientId := r.URL.Query().Get("client_id")
		if clientId == "" {
			http.Error(w, "missing parameter 'clientId'", http.StatusBadRequest)
			return
		}

		redirectUri := r.URL.Query().Get("redirect_uri")
		if redirectUri == "" {
			http.Error(w, "missing parameter 'redirectUri'", http.StatusBadRequest)
			return
		}

		scope := r.URL.Query().Get("scope")
		if scope == "" {
			http.Error(w, "missing parameter 'scope'", http.StatusBadRequest)
			return
		}

		state := getRandom()

		currentRequest = &authRequest{
			state:       state,
			clientId:    clientId,
			redirectUri: redirectUri,
		}

		http.Redirect(w, r, getUri(config, clientId, scope, state), http.StatusSeeOther)
	}
}

func CodeCallback(w http.ResponseWriter, r *http.Request) {

	if currentRequest == nil {
		http.Error(w, "unexpected callback", http.StatusBadRequest)
		return
	}
	defer func() { currentRequest = nil }()

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "missing parameter 'state'", http.StatusBadRequest)
		return
	}
	if state != currentRequest.state {
		http.Error(w, "unexpected callback 'state'", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing parameter 'code'", http.StatusBadRequest)
		return
	}

	url, err := url.Parse(currentRequest.redirectUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := url.Query()
	params.Add("auth_code", code)
	url.RawQuery = params.Encode()

	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func getRandom() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func GetRedirectUri(public_hostName string) string {
	return fmt.Sprintf("https://%v/auth/callback", public_hostName)
}

func getUri(config config.Config, clientId string, scope string, state string) string {

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {clientId},
		"redirect_uri":  {GetRedirectUri(config.PublicHostname)},
		"scope":         {scope},
		"state":         {state},
	}
	x := urlAuthorize + "?" + params.Encode()
	return x
}
