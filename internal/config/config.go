package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	MagicModulesPath string `json:"magicModulesPath"`
	GooglePath       string `json:"googlePath"`
	GoogleBetaPath   string `json:"googleBetaPath"`
	Remote           string `json:"remote"`
}

var CONFIG_FILE_NAME = ".tpg-cli-config.json"

func (c *Config) validate() []error {
	var errs []error
	if c.MagicModulesPath == "" {
		errs = append(errs, errors.New("error in loaded config: magicModulesPath is empty/missing"))
	} else {
		_, err := os.ReadDir(c.MagicModulesPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("error opening magicModulesPath path: %w", err))
		}
	}

	if c.GooglePath == "" {
		errs = append(errs, errors.New("error in loaded config: googlePath is empty/missing"))
	} else {
		_, err := os.ReadDir(c.GooglePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("error opening googlePath path: %w", err))
		}
	}

	if c.GoogleBetaPath == "" {
		errs = append(errs, errors.New("error in loaded config: googleBetaPath is empty/missing"))
	} else {
		_, err := os.ReadDir(c.GoogleBetaPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("error opening googleBetaPath path: %w", err))
		}
	}

	if c.Remote == "" {
		errs = append(errs, errors.New("error in loaded config: remote is empty/missing"))
	}
	return errs
}

func LoadConfigFromFile() (*Config, []error) {

	var errs []error
	home := os.Getenv("HOME")
	if home == "" {
		return nil, append(errs, errors.New("cannot find HOME environment variable, please make sure it is available"))
	}

	path := fmt.Sprintf("%s/%s", home, CONFIG_FILE_NAME)
	f, err := os.Open(path)
	if err != nil {
		return nil, append(errs, fmt.Errorf("error opening ~/%s: %w", CONFIG_FILE_NAME, err))
	}

	jsonParser := json.NewDecoder(f)
	config := Config{}
	if err = jsonParser.Decode(&config); err != nil {
		return nil, append(errs, fmt.Errorf("error parsing config file: %w", err))
	}

	errs = config.validate()
	if len(errs) > 0 {
		return nil, append(errs, fmt.Errorf("encountered %d errors when validating the contents of ~/%s", len(errs), CONFIG_FILE_NAME))
	}

	return &config, errs
}
