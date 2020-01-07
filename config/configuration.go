package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/lancer-kit/noble"
)

type Config struct {
	DB       string `json:"db"`
	LogLevel string `json:"log_level"`

	AuthLog     string    `json:"auth_log"`
	IgnoreFails bool      `json:"ignore_fails"`
	TG          *TGConfig `json:"tg,omitempty"`
}
type TGConfig struct {
	APIToken     noble.Secret        `json:"api_token"`
	AllowedUsers map[string]struct{} `json:"allowed_users"`
}

const (
	pathToLog = "/var/log/auth.log"
)

func GetConfig(configPath string) (config Config) {
	cfgFile, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = json.NewDecoder(cfgFile).Decode(&config)
	if err != nil {
		log.Fatal("Unable to read config:", err)
		return
	}
	if config.AuthLog == "" {
		config.AuthLog = pathToLog
	}

	if config.TG != nil {
		err = noble.RequiredSecret.Validate(config.TG.APIToken)
		if err != nil {
			log.Fatal("Secret Error:", err)
			return
		}
	}

	return
}
