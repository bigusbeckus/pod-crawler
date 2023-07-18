package config

import (
	_ "embed"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v3"
)

//go:embed config.yaml
var data []byte

// Use a single instance of validate since it caches struct info
// Reference: https://github.com/go-playground/validator/blob/master/_examples/simple/main.go#L27
var v *validator.Validate

type Config struct {
	Database struct {
		DBName   string `yaml:"dbName" validate:"required"`
		Host     string `yaml:"host" default:"http://localhost" validate:"required"`
		Password string `yaml:"password" validate:"required"`
		Port     uint16 `yaml:"port" default:"80" validate:"required"`
		User     string `yaml:"user" default:"postgres" validate:"required"`
	} `yaml:"database" validate:"required"`
	Input struct {
		PodcastListFile string `yaml:"podcastListFile" default:"data/podcasts.txt" validate:"required"`
	} `yaml:"input" validate:"required"`
	LogDestination string `yaml:"logDestination" default:"logs/" validate:"required"`
}

var AppConfig *Config

func Load() error {
	config := &Config{}

	// Set defaults
	if err := defaults.Set(config); err != nil {
		return err
	}

	// Unmarshal yaml
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Validate
	v = validator.New()
	if err := v.Struct(config); err != nil {
		return err
	}

	AppConfig = config
	return nil
}
