package breakerclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/breaker/breakerclient"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
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
}
