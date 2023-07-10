package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("no environment", func(t *testing.T) {
		t.Cleanup(func() {
			os.Remove("no_env.yaml")
		})
		require.NoError(t, NewConfig("no_env.yaml"))
		cfg, err := LoadConfig("no_env.yaml")
		require.NoError(t, err)
		require.Equal(t, "cosmoshub-4", cfg.Compass.ChainID)
		require.Equal(t, "tcp://127.0.0.1:26657", cfg.Compass.RPCAddr)
		require.Equal(t, "127.0.0.1:9090", cfg.Compass.GRPCAddr)
		logger, err := cfg.ZapLogger(true)
		require.NoError(t, err)
		logger.Debug("hello world")
		logger.Sync()
		apiOpts := cfg.ApiOpts(true)
		require.Equal(t, cfg.API.ListenAddress, apiOpts.ListenAddress)
		require.Equal(t, "", apiOpts.IdentifierField)
		require.Equal(t, cfg.API.TokenValidityDurationSeconds, apiOpts.TokenValidityDurationSeconds)
		require.Equal(t, true, apiOpts.DryRun)
	})
	t.Run("osmosis", func(t *testing.T) {
		t.Cleanup(func() {
			os.Remove("osmosis.yaml")
		})
		require.NoError(t, NewConfig("osmosis.yaml", "osmosis", "./osmosis_data"))
		cfg, err := LoadConfig("osmosis.yaml")
		require.NoError(t, err)
		require.Equal(t, "osmosis-1", cfg.Compass.ChainID)
		require.Equal(t, "https://osmosis-1.technofractal.com:443", cfg.Compass.RPCAddr)
		require.Equal(t, "https://gprc.osmosis-1.technofractal.com:443", cfg.Compass.GRPCAddr)
		logger, err := cfg.ZapLogger(true)
		require.NoError(t, err)
		logger.Debug("hello world")
		logger.Sync()
		apiOpts := cfg.ApiOpts(false)
		require.Equal(t, cfg.API.ListenAddress, apiOpts.ListenAddress)
		require.Equal(t, "osmosis", apiOpts.IdentifierField)
		require.Equal(t, cfg.API.TokenValidityDurationSeconds, apiOpts.TokenValidityDurationSeconds)
		require.Equal(t, false, apiOpts.DryRun)
	})
	t.Run("cosmos", func(t *testing.T) {
		t.Cleanup(func() {
			os.Remove("cosmos.yaml")
		})
		require.NoError(t, NewConfig("cosmos.yaml", "cosmos", "./cosmos_data"))
		cfg, err := LoadConfig("cosmos.yaml")
		require.NoError(t, err)
		require.Equal(t, "cosmoshub-4", cfg.Compass.ChainID)
		require.Equal(t, "https://cosmoshub-4.technofractal.com:443", cfg.Compass.RPCAddr)
		require.Equal(t, "https://gprc.cosmoshub-4.technofractal.com:443", cfg.Compass.GRPCAddr)
		logger, err := cfg.ZapLogger(true)
		require.NoError(t, err)
		logger.Debug("hello world")
		logger.Sync()
		apiOpts := cfg.ApiOpts(false)
		require.Equal(t, cfg.API.ListenAddress, apiOpts.ListenAddress)
		require.Equal(t, "cosmos", apiOpts.IdentifierField)
		require.Equal(t, cfg.API.TokenValidityDurationSeconds, apiOpts.TokenValidityDurationSeconds)
		require.Equal(t, false, apiOpts.DryRun)

	})
}
