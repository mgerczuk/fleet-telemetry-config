package tesla_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// https://developer.tesla.com/docs/fleet-api/authentication/third-party-tokens

const urlToken = "https://fleet-auth.prd.vn.cloud.tesla.com/oauth2/v3/token"

type FleetToken struct {
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresIn    int       `json:"expires_in"`
	TokenType    string    `json:"token_type"`
}

func getToken(params url.Values) (result *FleetToken, err error) {

	req, err := http.NewRequest("POST", urlToken, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(bodyBytes))
	}

	r1 := FleetToken{CreatedAt: time.Now()}
	err = json.Unmarshal(bodyBytes, &r1)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &r1, nil

}

func GetClientCredentials(clientId string, clientSecret string, audience string, scope string) (result *FleetToken, err error) {

	params := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"audience":      {audience},
		"scope":         {scope},
	}

	return getToken(params)
}

func GetAuthorizationCode(clientId string, clientSecret string, audience string, code string, redirectUrl string) (result *FleetToken, err error) {

	params := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"audience":      {audience},
		"code":          {code},
		"redirect_uri":  {redirectUrl},
	}

	return getToken(params)
}

func RefreshToken(clientId string, refreshToken string) (result *FleetToken, err error) {

	params := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientId},
		"refresh_token": {refreshToken},
	}

	return getToken(params)
}
