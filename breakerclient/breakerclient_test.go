package breakerclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/breaker/breakerclient"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

const (
	preExistingMnemonic = "muffin wrap reason cage blur crater uphold august silver slide loan home tag print this kiwi reflect run era cliff reveal minute bread garage"
)

func TestBreakerClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	cfg := compass.GetSimdConfig()
	breaker, err := breakerclient.NewBreakerClient(ctx, logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, breaker)
	disabledCmds, err := breaker.ListDisabledCommands(ctx)
	require.NoError(t, err)
	require.Len(t, disabledCmds.DisabledList, 0)

	mnemonic, err := breaker.NewMnemonic("example1")
	require.NoError(t, err)
	require.True(t, mnemonic != "")
	mnemonic, err = breaker.NewMnemonic("preExisting", preExistingMnemonic)
	require.NoError(t, err)
	require.Equal(t, mnemonic, preExistingMnemonic)
}
