package tesla_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// https://developer.tesla.com/docs/fleet-api/endpoints/partner-endpoints

const urlPartnerAccounts = "/api/1/partner_accounts"

type RegisterBody struct {
	Response         RegisterResponse `json:"response"`
	Error            string           `json:"error"`
	ErrorDescription string           `json:"error_description"`
	Txid             string           `json:"txid"`
}

type RegisterResponse struct {
	AccountID      string    `json:"account_id"`
	Ca             any       `json:"ca"`
	ClientID       string    `json:"client_id"`
	CreatedAt      time.Time `json:"created_at"`
	Csr            any       `json:"csr"`
	CsrUpdatedAt   any       `json:"csr_updated_at"`
	Description    string    `json:"description"`
	Domain         string    `json:"domain"`
	EnterpriseTier string    `json:"enterprise_tier"`
	Issuer         any       `json:"issuer"`
	Name           string    `json:"name"`
	PublicKey      string    `json:"public_key"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func Register(baseUrl string, token string, domain string) (result *RegisterResponse, err error) {

	bodyObj := map[string]interface{}{"domain": domain}
	js, _ := json.Marshal(bodyObj)

	req, err := http.NewRequest("POST", baseUrl+urlPartnerAccounts, bytes.NewReader(js))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

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

	var r RegisterBody
	err = json.Unmarshal(bodyBytes, &r)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &r.Response, nil
}
