package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/breaker/breakerclient"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

const (
	preExistingMnemonic = "muffin wrap reason cage blur crater uphold august silver slide loan home tag print this kiwi reflect run era cliff reveal minute bread garage"
)

func TestAPI(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("keyring-test")
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	cfg := compass.GetSimdConfig()
	breaker, err := breakerclient.NewBreakerClient(ctx, logger, cfg)
	require.NoError(t, err)

	// this part is needed in test environments since the key is not loaded
	_, err = breaker.NewMnemonic("default", preExistingMnemonic)
	require.NoError(t, err)

	jwt := NewJWT("password123", "userId", 3000)
	require.NoError(t, ConfigBreakerClient(breaker, "default"))
	api, err := NewAPI(ctx, logger, jwt, ApiOpts{
		ListenAddress:                "127.0.0.1:42690",
		Password:                     "password123",
		IdentifierField:              "userId",
		TokenValidityDurationSeconds: 3000,
	}, breaker)
	require.NoError(t, err)
	go func() {
		api.Serve()
	}()
	time.Sleep(time.Second * 2)
	api.logger.Info("issueing token")
	jwtToken, err := api.jwt.Encode("apiTest", nil)
	require.NoError(t, err)
	api.logger.Info("issued token", zap.String("token", jwtToken))
	// validate that the test environment setup process worked
	list, err := api.breakerClient.Accounts(ctx)
	require.NoError(t, err)
	require.True(t, len(list.Accounts) > 0)
	apiClient := NewAPIClient("http://127.0.0.1:42690", jwtToken)
	t.Run("v1", func(t *testing.T) {
		t.Run("webhook/mode_trip", func(t *testing.T) {
			resp, err := apiClient.TripCircuit([]string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"}, "amount > 1000")
			require.NoError(t, err)
			require.Equal(t, resp.Operation, MODE_TRIP)
			require.Equal(t, resp.Message, "ok")
			require.Equal(t, resp.Urls, []string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"})
			require.NotEmpty(t, resp.TxHash)
			t.Log(resp)
		})
		t.Run("webhook/unsupported_mode", func(t *testing.T) {
			apiClient.testRequestInvalidMode(t, []string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"}, "amount > 1000")
		})

		t.Run("status", func(t *testing.T) {
			t.Run("list_accounts", func(t *testing.T) {
				resp, err := apiClient.Accounts()
				require.NoError(t, err)
				require.True(t, len(resp.Accounts) > 0)
				t.Log(resp)
			})
			t.Run("list_disabled_commands", func(t *testing.T) {
				resp, err := apiClient.DisabledCommands()
				require.NoError(t, err)
				require.True(t, resp.DisabledList[0] == "/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker")
				require.True(t, len(resp.DisabledList) > 0)
				t.Log(resp)
			})
		})
		t.Run("webhook/mode_reset", func(t *testing.T) {
			resp, err := apiClient.ResetCircuit([]string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"}, "amount > 1000")
			require.NoError(t, err)
			require.Equal(t, resp.Operation, MODE_RESET)
			require.Equal(t, resp.Message, "ok")
			require.Equal(t, resp.Urls, []string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"})
			require.NotEmpty(t, resp.TxHash)
			t.Log(resp)
		})

		t.Run("status", func(t *testing.T) {
			t.Run("list_disabled_commands", func(t *testing.T) {
				resp, err := apiClient.DisabledCommands()
				require.NoError(t, err)
				require.True(t, len(resp.DisabledList) == 0)
				t.Log(resp)
			})
		})
	})
	api.Close()
}

// helper function used for testing
func (ac *APIClient) testRequestInvalidMode(t *testing.T, urls []string, message string) {
	payload := PayloadV1{
		Urls:      urls,
		Message:   message,
		Operation: Mode(100),
	}
	data, err := json.Marshal(&payload)
	if err != nil {
		panic(err)
	}
	buffer := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/webhook", ac.url), buffer)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", ac.jwt))
	res, err := ac.hc.Do(req)
	if err != nil {
		panic(err)
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(string(data), "unsupported mode") {
		panic(fmt.Sprintf("test failed %s", string(data)))
	}
}
