package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"cosmossdk.io/x/circuit/types"
)

type APIClient struct {
	hc  *http.Client
	url string
	jwt string
}

// Returns a new client for usage with the breaker api.
// Requires providing a valid JWT that has been issued, which can be done via the cli
//
// NOTE: JWT is not required for the `/status` api calls
//
// TODO: add a way of acquiring/renewing JWT via api
func NewAPIClient(url string, jwt string) APIClient {
	return APIClient{
		hc:  http.DefaultClient,
		url: url,
		jwt: jwt,
	}
}

// Returns all commands which have had a circuit tripped
func (ac *APIClient) DisabledCommands() (*types.DisabledListResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/status/list/disabledCommands", ac.url), &bytes.Buffer{})
	if err != nil {
		return nil, fmt.Errorf("failed to construct http request %s", err)
	}
	res, err := ac.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request %s", err)
	}
	var resp types.DisabledListResponse
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body %s", err)
	}
	if err = resp.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to deserialize http response body %s", err)
	}
	return &resp, nil
}

// Returns all accounts that have been granted some form of permission with the circuit breaker module
func (ac *APIClient) Accounts() (*types.AccountsResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/status/list/accounts", ac.url), &bytes.Buffer{})
	if err != nil {
		return nil, fmt.Errorf("failed to construct http request %s", err)
	}
	res, err := ac.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request %s", err)
	}
	var resp types.AccountsResponse
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body %s", err)
	}
	if err = resp.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("failed to deserialize http response body %s", err)
	}
	return &resp, nil
}

// Trips a circuit, preventing access to the given urls, emitting the `message` via system logs
func (ac *APIClient) TripCircuit(urls []string, message string) (*Response, error) {
	payload := PayloadV1{
		Urls:      urls,
		Message:   message,
		Operation: MODE_TRIP,
	}
	data, err := json.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload %s", err)
	}
	buffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/webhook", ac.url), buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to construct http request %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", ac.jwt))
	res, err := ac.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request %s", err)
	}
	return ac.unmarshalResponse(res.Body)
}

// Resets a circuit, allowing access to the given urls, emitting the `message` via system logs
func (ac *APIClient) ResetCircuit(urls []string, message string) (*Response, error) {
	payload := PayloadV1{
		Urls:      urls,
		Message:   message,
		Operation: MODE_RESET,
	}
	data, err := json.Marshal(&payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload %s", err)
	}
	buffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/webhook", ac.url), buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to construct http request %s", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", ac.jwt))
	res, err := ac.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request %s", err)
	}
	return ac.unmarshalResponse(res.Body)
}

func (ac *APIClient) unmarshalResponse(body io.ReadCloser) (*Response, error) {
	var resp Response

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body %s", err)
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to deserialize http response body %s", err)
	}
	return &resp, nil
}
