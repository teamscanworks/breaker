package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAPI(t *testing.T) {
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
	api.logger.Info("executing webhook")
	client := http.DefaultClient
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

	api.logger.Info("sleeping")
	time.Sleep(time.Second * 5)
	api.Close()
}
