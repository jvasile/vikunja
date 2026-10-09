package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// DisableRegistration closes public registration for the running process and
// persists the setting in config.yml. Existing YAML content is parsed and
// retained; missing mappings are created as needed.
func DisableRegistration() error {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		configPath = ResolvePath("config.yml")
	}

	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	values := map[string]any{}
	if len(content) > 0 {
		if err := yaml.Unmarshal(content, &values); err != nil {
			return err
		}
	}
	service, ok := values["service"].(map[string]any)
	if !ok {
		service = map[string]any{}
		values["service"] = service
	}
	service["enableregistration"] = false

	updated, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(configPath, updated, 0600); err != nil {
		return err
	}

	ServiceEnableRegistration.Set(false)
	return nil
}
