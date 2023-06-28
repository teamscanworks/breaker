package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFile(t *testing.T) {
	t.Cleanup(func() {
		os.Remove("config.yaml")
		os.Remove("/tmp/breaker.log")
	})
	err := NewConfig("config.yaml")
	require.NoError(t, err)
	_, err = LoadConfig("config.yyy")
	require.Error(t, err)
	cfg, err := LoadConfig("config.yaml")
	require.NoError(t, err)
	require.Equal(t, cfg.MonitoringAPI.ListenAddress, "http://127.0.0.1:26657")
	require.Equal(t, cfg.Cosmos.PrivateKey, "http://127.0.0.1:1317")
	logger, err := cfg.ZapLogger()
	require.NoError(t, err)
	logger.Info("hello world")
	logger.Sync()
}
