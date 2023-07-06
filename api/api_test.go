package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		api.logger.Info("executing webhook")
		payload := Payload{
			Urls:    []string{"/cosmos/apiv1"},
			Message: "amount > 1000",
		}
		data, err := json.Marshal(&payload)
		require.NoError(t, err)
		buffer := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", "http://127.0.0.1:42690/v1/webhook", buffer)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", jwtToken))
		require.NoError(t, err)
		res, err := client.Do(req)
		require.NoError(t, err)
		data, err = ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.Equal(t, string(data), "dry run, skipping transaction invocation")
	})
	api.logger.Info("sleeping")
	time.Sleep(time.Second * 5)
	api.Close()
}

func TestAPI(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("keyring-test")
	})
	t.Log("TODO: synchronize keyring with the simd environment")
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
	go func() {
		api.Serve()
	}()
	api.logger.Info("issueing token")
	jwtToken, err := api.jwt.Encode("apiTest", nil)
	require.NoError(t, err)
	api.logger.Info("issued token", zap.String("token", jwtToken))
	client := http.DefaultClient

	t.Run("v1/webhook", func(t *testing.T) {
		api.logger.Info("executing webhook")
		payload := Payload{
			Urls:    []string{"/cosmos/apiv1"},
			Message: "amount > 1000",
		}
		data, err := json.Marshal(&payload)
		require.NoError(t, err)
		buffer := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", "http://127.0.0.1:42690/v1/webhook", buffer)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %s", jwtToken))
		require.NoError(t, err)
		res, err := client.Do(req)
		require.NoError(t, err)
		data, err = ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		_ = data
	})
	api.logger.Info("sleeping")
	time.Sleep(time.Second * 5)
	api.Close()
}
