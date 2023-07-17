package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func TestAPIDryRun(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("keyring-test")
	})
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	jwt := NewJWT("password123", "userId", 300)

	api, err := NewAPI(ctx, logger, jwt, ApiOpts{
		ListenAddress:                "127.0.0.1:42690",
		Password:                     "password123",
		IdentifierField:              "userId",
		TokenValidityDurationSeconds: 300,
		DryRun:                       true,
	})
	require.NoError(t, err)
	go func() {
		api.Serve()
	}()
	api.logger.Info("issueing token")
	jwtToken, err := api.jwt.Encode("apiTest", nil)
	require.NoError(t, err)
	api.logger.Info("issued token", zap.String("token", jwtToken))
	client := http.DefaultClient

	t.Run("v1/webhook", func(t *testing.T) {
		t.Run("mode_reset", func(t *testing.T) {

			api.logger.Info("executing webhook")
			payload := PayloadV1{
				Urls:      []string{"/cosmos/apiv1"},
				Message:   "amount > 1000",
				Operation: MODE_RESET,
			}
			data, err := json.Marshal(&payload)
			require.NoError(t, err)
			buffer := bytes.NewBuffer(data)
			req, err := http.NewRequest("POST", "http://127.0.0.1:42690/v1/webhook", buffer)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", jwtToken))
			require.NoError(t, err)
			res, err := client.Do(req)
			require.NoError(t, err)
			data, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, string(data), "dry run, skipping transaction invocation")

		})
		t.Run("mode_trip", func(t *testing.T) {

			api.logger.Info("executing webhook")
			payload := PayloadV1{
				Urls:      []string{"/cosmos/apiv1"},
				Message:   "amount > 1000",
				Operation: MODE_TRIP,
			}
			data, err := json.Marshal(&payload)
			require.NoError(t, err)
			buffer := bytes.NewBuffer(data)
			req, err := http.NewRequest("POST", "http://127.0.0.1:42690/v1/webhook", buffer)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", jwtToken))
			require.NoError(t, err)
			res, err := client.Do(req)
			require.NoError(t, err)
			data, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, string(data), "dry run, skipping transaction invocation")

		})
	})
	api.logger.Info("sleeping")
	time.Sleep(time.Second * 5)
	api.Close()
}

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
	jwt := NewJWT("password123", "userId", 300)

	api, err := NewAPI(ctx, logger, jwt, ApiOpts{
		ListenAddress:                "127.0.0.1:42690",
		Password:                     "password123",
		IdentifierField:              "userId",
		TokenValidityDurationSeconds: 300,
		DryRun:                       false,
	})
	require.NoError(t, err)
	api.WithBreakerClient(breaker)
	_, err = api.breakerClient.NewMnemonic("default", preExistingMnemonic)
	require.NoError(t, err)
	require.NoError(t, api.breakerClient.SetFromAddress())
	api.breakerClient.UpdateClientFromName("default")
	go func() {
		api.Serve()
	}()
	time.Sleep(time.Second * 2)
	api.logger.Info("issueing token")
	jwtToken, err := api.jwt.Encode("apiTest", nil)
	require.NoError(t, err)
	api.logger.Info("issued token", zap.String("token", jwtToken))
	client := http.DefaultClient
	// validate that the test environment setup process worked
	list, err := api.breakerClient.Accounts(ctx)
	require.NoError(t, err)
	require.True(t, len(list.Accounts) > 0)
	t.Run("v1/webhook", func(t *testing.T) {
		t.Run("mode_trip", func(t *testing.T) {
			api.logger.Info("executing webhook")
			payload := PayloadV1{
				Urls:      []string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"},
				Message:   "amount > 1000",
				Operation: MODE_TRIP,
			}
			data, err := json.Marshal(&payload)
			require.NoError(t, err)
			buffer := bytes.NewBuffer(data)
			req, err := http.NewRequest("POST", "http://127.0.0.1:42690/v1/webhook", buffer)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", jwtToken))
			require.NoError(t, err)
			res, err := client.Do(req)
			require.NoError(t, err)
			data, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			// TODO: deserialize and validate response values
			t.Log(string(data))
		})
		t.Run("mode_reset", func(t *testing.T) {
			payload := PayloadV1{
				Urls:      []string{"/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"},
				Message:   "amount > 1000",
				Operation: MODE_RESET,
			}
			data, err := json.Marshal(&payload)
			require.NoError(t, err)
			buffer := bytes.NewBuffer(data)
			req, err := http.NewRequest("POST", "http://127.0.0.1:42690/v1/webhook", buffer)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", jwtToken))
			require.NoError(t, err)
			res, err := client.Do(req)
			require.NoError(t, err)
			data, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			// TODO: deserialize and validate response values
			t.Log(string(data))
		})
	})
	api.logger.Info("sleeping")
	time.Sleep(time.Second * 5)
	api.Close()
}
