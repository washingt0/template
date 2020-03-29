package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var cfg *Config

// Config contains all needed settings
type Config struct {
	ApplicationName string   `json:"application_name"`
	RWDatabase      string   `json:"rw_database"`
	RODatabase      string   `json:"ro_database"`
	LogPath         string   `json:"log_path"`
	HTTPAddress     string   `json:"http_address"`
	GRPCAddress     string   `json:"grpc_address"`
	Production      bool     `json:"production"`
	Secrets         []string `json:"secrets"`
}

// GetConfig returns a config copy
func GetConfig() Config {
	if cfg == nil {
		cfg = &Config{
			ApplicationName: "template",
			RWDatabase:      "postgres://postgres:123456@localhost:5432/template",
			LogPath:         "/tmp/log/template/",
			Secrets:         []string{"123456"},
		}
	}

	return *cfg
}

// LoadConfig load config from a json config file or environment
func LoadConfig() (err error) {
	var (
		cfgPath string = "config.json"
		data    []byte
	)

	if val := os.Getenv("APPLICATION_CONFIG"); val != "" {
		cfgPath = val
	}

	if data, err = ioutil.ReadFile(cfgPath); err != nil {
		return
	}

	cfg = new(Config)

	if err = json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	return
}
