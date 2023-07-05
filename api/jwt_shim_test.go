package api_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/breaker/api"
)

func TestJwt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type test struct {
		name                     string
		validPassword            string
		identifierField          string
		identifierFieldValue     string
		tokenValidityDurationSec int64
		wantErr                  bool
	}
	tests := []test{
		{
			name:                     "auth.ok",
			validPassword:            "password123",
			identifierField:          "user_id",
			identifierFieldValue:     "validUser",
			tokenValidityDurationSec: 300,
			wantErr:                  false,
		},
		{
			name:                     "auth.fail_expired",
			validPassword:            "password123",
			identifierField:          "user_id",
			identifierFieldValue:     "validUser",
			tokenValidityDurationSec: 1,
			wantErr:                  true, // fails due to token expiration
		},
		{
			name:                     "auth.fail_badlookup",
			validPassword:            "password123",
			identifierField:          "foobar",
			identifierFieldValue:     "",
			tokenValidityDurationSec: 300,
			wantErr:                  true, // fails due to invalid field to check
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := api.NewJWT(tt.validPassword, tt.identifierField, tt.tokenValidityDurationSec)

			encodedJwt, err := jwt.Encode(tt.identifierFieldValue, nil)
			require.NoError(t, err)
			time.Sleep(time.Second * 5)
			decodedJwt, err := jwt.Decode(ctx, encodedJwt)
			require.NoError(t, err)
			fmt.Println(decodedJwt)
			err = jwt.CheckToken(decodedJwt)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

		})
	}
}
