package breakerclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamscanworks/breaker/config"
)

func TestBreakerClientNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	cfg := config.ExampleConfig
	logger, err := cfg.ZapLogger()
	require.NoError(t, err)
	defer logger.Sync()
	bc, err := NewBreakerClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, bc)
	defer bc.Close()

	accounts, err := bc.Accounts(ctx)
	require.NoError(t, err)
	t.Logf("accounts %v", accounts)

	disabledCmds, err := bc.ListDisabledCommands(ctx)
	require.NoError(t, err)
	require.Len(t, disabledCmds.DisabledList, 0)

	require.NoError(t, bc.TripCircuitBreaker(ctx, []string{"foobar"}))

	disabledCmds, err = bc.ListDisabledCommands(ctx)
	require.NoError(t, err)
	require.Len(t, disabledCmds.DisabledList, 1)
}
