package config

import (
	"io/ioutil"
	"os"

	"github.com/99designs/keyring"
	"github.com/teamscanworks/breaker/api"
	"github.com/teamscanworks/compass"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var (
	ExampleConfig = Configuration{
		Compass: *compass.GetSimdConfig(),
		API: API{
			ListenAddress: "http://127.0.0.1:6666",
			Password:      "password123",
			// empty means no extra identifier is used
			IdentifierField:              "",
			TokenValidityDurationSeconds: 86400,
		},
	}
)

type Configuration struct {
	Compass compass.ClientConfig `yaml:"compass"`
	API     API                  `yaml:"api"`
}

type API struct {
	ListenAddress string `yaml:"listen_address"`
	// password for the JWT
	Password string `yaml:"password"`
	// field used to store additional information in
	IdentifierField string `yaml:"identifier_field"`
	// time in seconds that issued jwt's are valid for
	TokenValidityDurationSeconds int64 `yaml:"token_validity_duration_seconds"`
}

func NewConfig(path string) error {
	data, err := yaml.Marshal(&ExampleConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.ModePerm)
}

func LoadConfig(path string) (*Configuration, error) {
	var (
		r   []byte
		cfg Configuration
	)
	r, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(r, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// make sure to call `logger.Sync()` at somepoint before exiting
func (c *Configuration) ZapLogger(debug bool) (*zap.Logger, error) {
	conf := zap.NewProductionConfig()
	if debug {
		conf.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		keyring.Debug = true
	}
	logger, err := zap.NewProduction()
	return logger, err
}

func (c *Configuration) ApiOpts(dryRun bool) api.ApiOpts {
	return api.ApiOpts{
		ListenAddress:                c.API.ListenAddress,
		IdentifierField:              c.API.IdentifierField,
		TokenValidityDurationSeconds: c.API.TokenValidityDurationSeconds,
		DryRun:                       dryRun,
	}
}
