package config

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/teamscanworks/breaker/api"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var (
	// an example config suitable for testing
	ExampleConfig = Configuration{
		Compass: *compass.GetSimdConfig(),
		API: API{
			ListenAddress: "http://127.0.0.1:6666",
			Password:      "password123",
			// empty means no extra identifier is used when validating jwts
			IdentifierField:              "",
			TokenValidityDurationSeconds: 86400,
		},
	}
)

// exposes compass client configuration, as well as the breaker api config
type Configuration struct {
	Compass compass.ClientConfig `yaml:"compass"`
	API     API                  `yaml:"api"`
}

// configures the breaker api
type API struct {
	// address that the api is served on
	ListenAddress string `yaml:"listen_address"`
	// password used for encoding/decoding and verifying jwts
	Password string `yaml:"password"`
	// field used to store additional information in
	// can be left empty if not needed
	IdentifierField string `yaml:"identifier_field"`
	// time in seconds that issued jwt's are valid for
	TokenValidityDurationSeconds int64 `yaml:"token_validity_duration_seconds"`
}

// Saves the example configuration at `path` as a yaml file, you may
// override the default configuration which is suitable for simd to one
// suitable for cosmos, or osmosis by supplying two values for `environment`.
//
// When containing two elements, environment[0] is the network to adjust the
// configuration for, currently supporting "cosmos" and "osmosis", while environment[1]
// is the path to use as the "home" directory, namely for keyring storage.
func NewConfig(path string, environment ...string) error {
	// set the default configuration, usable in simd environments
	cfg := ExampleConfig
	if len(environment) > 0 {
		fmt.Println("environment", environment)
		if environment[0] == "cosmos" && len(environment) == 2 {
			cfg.Compass = *compass.GetCosmosHubConfig(environment[1], true)
			cfg.API.IdentifierField = "cosmos"
		} else if environment[0] == "osmosis" && len(environment) == 2 {
			cfg.Compass = *compass.GetOsmosisConfig(environment[1], true)
			cfg.API.IdentifierField = "osmosis"
		}
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, os.ModePerm)
}

// Loads the configuration from `path` which must be saved as yaml file.
func LoadConfig(path string) (*Configuration, error) {
	var (
		err error
		dat []byte
		cfg Configuration
	)
	dat, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(dat, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Returns an initialized zap production logger, optionally with debug logs enabled
// you must call `logger.Sync()` once sometime before the process exits although
// not doing is probably ok.
func (c *Configuration) ZapLogger(debug bool) (*zap.Logger, error) {
	conf := zap.NewProductionConfig()
	if debug {
		conf.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		keyring.Debug = true
	}
	logger, err := zap.NewProduction()
	return logger, err
}

// Returns an instance of the api options struct, which can be set to not broadcast
// any transactions by setting `dryRun` to true.
func (c *Configuration) ApiOpts(dryRun bool) api.ApiOpts {
	return api.ApiOpts{
		ListenAddress:                c.API.ListenAddress,
		IdentifierField:              c.API.IdentifierField,
		TokenValidityDurationSeconds: c.API.TokenValidityDurationSeconds,
		DryRun:                       dryRun,
	}
}
