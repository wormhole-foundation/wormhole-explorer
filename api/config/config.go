package config

import (
	"bytes"
	"encoding/json"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/spf13/viper"
	"strings"
)

type AppConfig struct {
	DB struct {
		URL string
		// database name
		Name string
	}

	PORT int

	LogLevel string
}

func (cfg *AppConfig) GetLogLevel() (ipfslog.LogLevel, error) {
	return ipfslog.LevelFromString(cfg.LogLevel)
}

func init() {
	viper.SetDefault("port", 8000)
	viper.SetDefault("loglevel", "INFO")
	// Consider environment variables in unmarshall doesn't work unless doing this: https://github.com/spf13/viper/issues/188#issuecomment-1168898503
	b, err := json.Marshal(AppConfig{})
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

func Get() (*AppConfig, error) {
	var cfg AppConfig
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}
