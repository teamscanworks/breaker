package config

import (
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

var (
	ExampleConfig = Config{
		Cosmos: Cosmos{
			RPCEndpoint:  "http://127.0.0.1:1317",
			GRPCEndpoint: "http://127.0.0.1:9090",
			PrivateKey:   "",
		},
		MonitoringAPI: MonitoringAPI{
			ListenAddress: "127.0.0.1:3070",
			TLS: TLS{
				CertPath: "",
				KeyPath:  "",
			},
		},
		Logger: Logger{
			Path:  "/tmp/breaker.log",
			Debug: true,
		},
	}
)

// Config provides the base configuration object used for breaker
type Config struct {
	Cosmos        Cosmos        `yaml:"cosmos"`
	MonitoringAPI MonitoringAPI `yaml:"monitoring_api"`
	Logger        Logger        `yaml:"logger"`
}

// Cosmos provides configuration of our connection to a cosmos chain
type Cosmos struct {
	// rpc endpoint we use to communicate with a cosmos chain
	RPCEndpoint  string `yaml:"rpc_endpoint"`
	GRPCEndpoint string `yaml:"grpc_endpoint"`
	// hex encoded private key used for signing transactions
	PrivateKey string     `yaml:"private_key"`
	Options    CosmosOpts `yaml:"options"`
}

type CosmosOpts struct {
	HomeDir    string `yaml:"home_dir"`
	ChainId    string `yaml:"chain_id"`
	KeyringDir string `yaml:"keyring_dir"`
}

// MonitoringAPI provides configuration of the monitoring api endpoint
// which exposes metric pushing functionality
type MonitoringAPI struct {
	ListenAddress string `yaml:"listen_address"`
	TLS           TLS    `yaml:"tls"`
}

// Logger provides configuration over zap logger
type Logger struct {
	// Path to store the configuration file
	Path string `yaml:"path"`
	// Debug enables displaying of debug logs
	Debug bool `yaml:"debug"`
}

// TLS provides configuration for enabling TLS secured connections
type TLS struct {
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
}

func NewConfig(path string) error {
	data, err := yaml.Marshal(&ExampleConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.ModePerm)
}

func LoadConfig(path string) (cfg *Config, err error) {
	var r []byte
	r, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(r, &cfg); err != nil {
		return nil, err
	}
	return
}

// returns an initialized zap.Logger
// make sure to call `logger.Sync()` at somepoint before exiting
func (c *Config) ZapLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	return logger, err
}

func (c *Config) ClientContext() client.Context {
	ctx := client.Context{}.WithViper("breaker")
	if c.Cosmos.Options.HomeDir != "" {
		ctx = ctx.WithHomeDir(c.Cosmos.Options.HomeDir)
	}
	if c.Cosmos.Options.ChainId != "" {
		ctx = ctx.WithChainID(c.Cosmos.Options.ChainId)
	}
	if c.Cosmos.Options.KeyringDir != "" {
		ctx = ctx.WithKeyringDir(c.Cosmos.Options.KeyringDir)
	}
	return ctx
}
