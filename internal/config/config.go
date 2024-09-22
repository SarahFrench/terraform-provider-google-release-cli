package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	MagicModulesPath string `json:"magicModulesPath"`
	GooglePath       string `json:"googlePath"`
	GoogleBetaPath   string `json:"googleBetaPath"`
	Remote           string `json:"remote"`
}

type compositeValidationError []error

func (ve compositeValidationError) Error() string {
	if len(ve) > 0 {
		var b strings.Builder
		b.WriteString("There were some problems with the command's config:\n")
		for _, e := range ve {
			b.WriteString(fmt.Sprintf("\t> %v\n", e))
		}
		return b.String()
	}
	return "this error should not be surfaced, and if it's observed it's due to a bug in the CLI"
}

var CONFIG_FILE_NAME = ".tpg-cli-config.json"

func (c *Config) validate() error {

	var errs compositeValidationError

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

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func LoadConfigFromFile() (*Config, error) {

	var errs compositeValidationError
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

	err = config.validate()
	if err != nil {
		return nil, append(errs, err)
	}

	if len(errs) > 0 {
		return nil, append(errs, err)
	}

	return &config, nil
}
