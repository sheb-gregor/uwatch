package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancer-kit/noble"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	pathToLog = "/var/log/auth.log"
)

func GetConfig(configPath string) Config {
	cfgFile, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return Config{}
	}

	var decoder interface {
		Decode(v interface{}) error
	}

	ext := filepath.Ext(cfgFile.Name())
	switch ext {
	case ".json":
		decoder = json.NewDecoder(cfgFile)
	case ".yaml":
		decoder = yaml.NewDecoder(cfgFile)
	default:
		log.Fatal("Invalid file extension:", ext)
		return Config{}
	}

	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Unable to read config:", err)
		return config
	}

	if config.AuthLog == "" {
		config.AuthLog = pathToLog
	}

	if config.TG != nil {
		err = noble.RequiredSecret.Validate(config.TG.APIToken)
		if err != nil {
			log.Fatal("Secret Error:", err)
			return config
		}
	}

	err = config.Init()
	if err != nil {
		log.Fatal("Unable to inti config:", err)
		return config
	}

	return config
}

type Config struct {
	Server      string `json:"server" yaml:"server"`
	AuthLog     string `json:"auth_log" yaml:"auth_log"`
	DataDir     string `json:"data_dir" yaml:"data_dir"`
	LogLevel    string `json:"log_level" yaml:"log_level"`
	IgnoreFails bool   `json:"ignore_fails" yaml:"ignore_fails"`

	TG *TGConfig `json:"tg,omitempty" yaml:"tg"`
}

func (cfg *Config) Init() error {
	_, err := os.Stat(cfg.DataDir)
	if os.IsNotExist(err) {
		return os.Mkdir(cfg.DataDir, 0755)
	}

	if cfg.Server == "" {
		cfg.Server, _ = os.Hostname()
	}

	return nil
}

func (cfg Config) Logger() *logrus.Entry {
	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.SetLevel(logLevel)

	return logger.WithField("app", "uwatch")
}

func (cfg Config) GetPath(ext string) string {
	return strings.TrimSuffix(cfg.DataDir, "/") + "/uwatch." + ext
}

type TGConfig struct {
	APIToken     noble.Secret        `json:"api_token" yaml:"api_token"`
	AllowedUsers map[string]struct{} `json:"allowed_users" yaml:"allowed_users"`
}
