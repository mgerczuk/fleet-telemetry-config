package tesla_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/teslamotors/vehicle-command/pkg/protocol"
)

// https://developer.tesla.com/docs/fleet-api/endpoints/vehicle-endpoints

type VehicleClient struct {
	baseURL string
	// vehicleTag string
	token string
	// HTTPClient *http.Client
}

func NewVehicleClient(baseUrl string, token string) VehicleClient {
	return VehicleClient{
		baseURL: baseUrl,
		// vehicleTag: vehicleTag,
		token: token,
		// HTTPClient: &http.Client{}
	}
}

// ==========================================================================

type FieldProp struct {
	IntervalSeconds       int      `json:"interval_seconds"`
	MinimumDelta          *float32 `json:"minimum_delta,omitempty"`
	ResendIntervalSeconds *int     `json:"resend_interval_seconds,omitempty"`
}

type FleetTelemetryConfigData struct {
	Port        int                  `json:"port"`
	Exp         *int                 `json:"exp,omitempty"`
	AlertTypes  []string             `json:"alert_types"`
	Fields      map[string]FieldProp `json:"fields"`
	Ca          string               `json:"ca,omitempty"`
	Hostname    string               `json:"hostname,omitempty"`
	PreferTyped *bool                `json:"prefer_typed,omitempty"`
	// read-only
	Aud *string `json:"aud,omitempty"`
	Iss *string `json:"iss,omitempty"`
}

type GetFleetTelemetryConfigResponse struct {
	Config FleetTelemetryConfigData `json:"config"`
	Synced bool                     `json:"synced"`
}

type getFleetTelemetryConfigResponseBody struct {
	Response         GetFleetTelemetryConfigResponse `json:"response"`
	Error            string                          `json:"error"`
	ErrorDescription string                          `json:"error_description"`
	Txid             string                          `json:"txid"`
}

func (c *VehicleClient) GetFleetTelemetryConfig(vehicleTag string) (*GetFleetTelemetryConfigResponse, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/1/vehicles/%s/fleet_telemetry_config", c.baseURL, vehicleTag), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	//util.LogResponseBody(res)

	var body getFleetTelemetryConfigResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, errors.New("failed to decode response json")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(body.Error)
	}

	return &body.Response, nil
}

// ==========================================================================

func toMapClient(config *FleetTelemetryConfigData) jwt.MapClaims {

	fields := make(map[string]interface{})
	for key, val := range config.Fields {
		v := map[string]interface{}{
			"interval_seconds": val.IntervalSeconds,
		}
		if val.MinimumDelta != nil {
			v["minimum_delta"] = val.MinimumDelta
		}
		if val.ResendIntervalSeconds != nil {
			v["resend_interval_seconds"] = val.ResendIntervalSeconds
		}
		fields[key] = v
	}

	mc := jwt.MapClaims{
		"hostname":    config.Hostname,
		"port":        config.Port,
		"ca":          config.Ca,
		"exp":         config.Exp,
		"alert_types": config.AlertTypes,
		"fields":      fields,
	}

	return mc
}

type SetFleetTelemetryConfigResponse struct {
	UpdatedVehicles int `json:"updated_vehicles"`
	SkippedVehicles struct {
		MissingKey          []any    `json:"missing_key"`
		UnsupportedHardware []string `json:"unsupported_hardware"`
		UnsupportedFirmware []string `json:"unsupported_firmware"`
	} `json:"skipped_vehicles"`
}

type setFleetTelemetryConfigResponseBody struct {
	Response         SetFleetTelemetryConfigResponse `json:"response"`
	Error            string                          `json:"error"`
	ErrorDescription string                          `json:"error_description"`
	Txid             string                          `json:"txid"`
}

func (c *VehicleClient) CreateFleetTelemetryConfig(config *FleetTelemetryConfigData, vins []string, privateKey string) (*SetFleetTelemetryConfigResponse, error) {

	mc := toMapClient(config)

	key, err := protocol.LoadStringECDHKey(privateKey)
	if err != nil {
		return nil, err
	}

	token, err := protocol.SignMessageForFleet(key, "TelemetryClient", mc)
	if err != nil {
		return nil, err
	}

	bodyObj := map[string]interface{}{
		"token": token,
		"vins":  vins,
	}
	js, _ := json.Marshal(bodyObj)

	req, err := http.NewRequest("POST", c.baseURL+"/api/1/vehicles/fleet_telemetry_config_jws", bytes.NewReader(js))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.token)

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

	var r setFleetTelemetryConfigResponseBody
	err = json.Unmarshal(bodyBytes, &r)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &r.Response, nil
}

// ==========================================================================

func (c *VehicleClient) DeleteFleetTelemetryConfig(vehicleTag string) (*SetFleetTelemetryConfigResponse, error) {

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/1/vehicles/%s/fleet_telemetry_config", c.baseURL, vehicleTag), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var body setFleetTelemetryConfigResponseBody
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return nil, errors.New("failed to decode response json")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(body.Error)
	}

	return &body.Response, nil
}
