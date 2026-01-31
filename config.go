package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json"), nil
}

func loadConfig(path string) (*Config, error) {
	config := &Config{
		Schema:   "https://opencode.ai/config.json",
		Provider: make(map[string]Provider),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.Provider == nil {
		config.Provider = make(map[string]Provider)
	}

	if config.MCP == nil {
		config.MCP = make(map[string]MCPServer)
	}

	return config, nil
}

func saveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
