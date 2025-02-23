// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type ConfigLoader interface {
	LoadConfiguration(ctx context.Context) (*AppConf, error)
}

type DefaultConfigLoader struct{}

// LoadConfiguration loads the application configuration from various sources.
func (DefaultConfigLoader) LoadConfiguration() (*AppConf, error) {
	appConfig := &AppConf{GommitConf: &Gommit{Conventional: &Conventional{}}}
	if err := ReadConfigurationFile(appConfig, ".gommitlint.yaml"); err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	defaultConventionalTypes(appConfig)

	//	for _, config := range appConfig.Configurations {
	// if err := validateConfiguration(config); err != nil {
	// 	return nil, fmt.Errorf("failed to validate configuration: %w", err)
	// }
	//	}

	//	fmt.Print(appConfig.CommitDeclaration)
	//	fmt.Print(appConfig)

	return appConfig, nil
}

func defaultConventionalTypes(appConf *AppConf) {
	// fill defaults
	if len(appConf.GommitConf.Conventional.Types) == 0 {
		types := []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}

		appConf.GommitConf.Conventional.Types = types
	}
}

func ReadConfigurationFile(appConfiguration *AppConf, configfile string) error {
	const (
		xdgConfigHomeEnv        = "XDG_CONFIG_HOME"
		xdgConfigHomeConfigPath = "/gommitlint/" + "gommitlint.yaml"
	)

	koanfConf := koanf.New(".")
	xdgConfigfileExists, xdgConfigFilePath := hasXDGConfigFile(xdgConfigHomeEnv, xdgConfigHomeConfigPath)
	localConfigfileExists := hasLocalConfigFile(configfile)

	// xdg config file
	if xdgConfigfileExists {
		if err := koanfConf.Load(file.Provider(xdgConfigFilePath), yaml.Parser()); err != nil {
			return fmt.Errorf("error loading xdg_config_home configuration. %w", err)
		}
	}

	// local config file
	if localConfigfileExists {
		if err := koanfConf.Load(file.Provider(configfile), yaml.Parser()); err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
	}

	if !localConfigfileExists && !xdgConfigfileExists {
		return nil
	}

	// Unmarshal the YAML data into the config struct
	if err := koanfConf.Unmarshal("", appConfiguration); err != nil {
		panic(fmt.Errorf("error unmarshalling yaml config: %w", err))
	}

	// if len(appConfiguration) == 0 {
	// 	panic("No configuration could be found!")
	// }

	return nil
}

func hasXDGConfigFile(xdgconfighome string, xdgconfighomeconfigpath string) (bool, string) {
	xdgConfigfileExists := false

	var xdgConfigFilePath string

	envValue, xdgHomeIsSet := os.LookupEnv(xdgconfighome)
	if xdgHomeIsSet {
		xdgConfigFilePath = filepath.Join(envValue, xdgconfighomeconfigpath)
		if _, err := os.Stat(xdgConfigFilePath); err == nil {
			xdgConfigfileExists = true
		}
	}

	return xdgConfigfileExists, xdgConfigFilePath
}

func hasLocalConfigFile(configFile string) bool {
	localConfigfileExists := false
	if _, err := os.Stat(configFile); err == nil {
		localConfigfileExists = true
	}

	return localConfigfileExists
}
