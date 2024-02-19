// Package config implement a simple configuration package.
// It define a type [AppConfig] that represent the aplication configuration and
// use viper [https://github.com/spf13/viper] to load the configuration.
package config

import (
	"bytes"
	"encoding/json"
	"strings"

	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/spf13/viper"
)

const (
	RunModeProduction   = "PRODUCTION"
	RunModeDevelopmernt = "DEVELOPMENT"
)

// p2p network constants.
const (
	P2pMainNet = "mainnet"
	P2pTestNet = "testnet"
	P2pDevNet  = "devnet"
)

// AppConfig defines the configuration for the app.
type AppConfig struct {
	DB struct {
		URL string
		// database name
		Name string
	}
	Cache struct {
		URL              string
		TvlKey           string
		TvlExpiration    int
		Enabled          bool
		MetricExpiration int
		Prefix           string
	}
	PORT         int
	LogLevel     string
	RunMode      string
	P2pNetwork   string
	PprofEnabled bool
	Environment  string
	Influx       struct {
		URL            string
		Token          string
		Organization   string
		Bucket24Hours  string
		Bucket30Days   string
		BucketInfinite string
	}
	VaaPayloadParser struct {
		Enabled bool
		URL     string
		Timeout int64
	}
	RateLimit struct {
		Enabled bool
		// Max number of requests per minute
		Max int
		// Prefix for redis keys
		Prefix string
		//Api Tokens
		Tokens string
	}
	Protocols                []string
	ProtocolsStatsVersion    string
	ProtocolsActivityVersion string
}

// GetLogLevel get zapcore.Level define in the configuraion.
func (cfg *AppConfig) GetLogLevel() (ipfslog.LogLevel, error) {
	return ipfslog.LevelFromString(cfg.LogLevel)
}

func defaulConfig() *AppConfig {
	return &AppConfig{
		Cache: struct {
			URL              string
			TvlKey           string
			TvlExpiration    int
			Enabled          bool
			MetricExpiration int
			Prefix           string
		}{
			MetricExpiration: 10,
		},
	}
}

func init() {
	viper.SetDefault("port", 8000)
	viper.SetDefault("loglevel", "INFO")
	viper.SetDefault("runmode", "PRODUCTION")
	viper.SetDefault("p2pnetwork", P2pMainNet)
	viper.SetDefault("PprofEnabled", false)
	viper.SetDefault("RateLimit_Enabled", true)

	// Consider environment variables in unmarshall doesn't work unless doing this: https://github.com/spf13/viper/issues/188#issuecomment-1168898503
	b, err := json.Marshal(defaulConfig())
	if err != nil {
		panic(err)
	}
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("yaml")
	if err := viper.MergeConfig(defaultConfig); err != nil {
		panic(err)
	}
	// overwrite values from config
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigParseError); ok {
			panic(err)
		}
		// dont return error if file is missing. overwrite file is optional
	}
	// tell viper to overrwire env variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("WORMSCAN")
	viper.AutomaticEnv()
}

// Get returns the app configuration.
func Get() (*AppConfig, error) {
	var cfg AppConfig
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}

func (c *AppConfig) GetApiTokens() []string {
	return strings.Split(c.RateLimit.Tokens, ",")
}
