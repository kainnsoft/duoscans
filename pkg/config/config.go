package config

import (
	"fmt"
	"os"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"

	"github.com/kainnsoft/duoscans/config"
)

const configPathKey = "APP_CONF_PATH"

// FromEnv creates a new config based on the file path in APP_CONF_PATH.
func FromEnv() (config.Config, error) {
	envFileName, ok := os.LookupEnv(configPathKey)
	if !ok {
		return config.Config{}, fmt.Errorf("variable %s is empty", configPathKey)
	}

	var cfg config.Config

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{envFileName},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	}).Load(); err != nil {
		return config.Config{}, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}
