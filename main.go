package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Schema            string              `json:"$schema"`
	Provider          map[string]Provider `json:"provider"`
	Model             string              `json:"model,omitempty"`
	SmallModel        string              `json:"small_model,omitempty"`
	EnabledProviders  []string            `json:"enabled_providers,omitempty"`
	DisabledProviders []string            `json:"disabled_providers,omitempty"`
}

type Provider struct {
	NPM     string                 `json:"npm"`
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
	Models  map[string]Model       `json:"models"`
}

type Model struct {
	Name  string      `json:"name"`
	ID    string      `json:"id,omitempty"`
	Limit *ModelLimit `json:"limit,omitempty"`
}

type ModelLimit struct {
	Context int `json:"context,omitempty"`
	Output  int `json:"output,omitempty"`
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "opencode", "opencode.json"), nil
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

	return config, nil
}

func saveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func promptString(prompt string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	var input string
	fmt.Scanln(&input)

	if input == "" {
		return defaultValue
	}
	return input
}

func promptBool(prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("%s [%s] (y/n): ", prompt, defaultStr)

	var input string
	fmt.Scanln(&input)

	if input == "" {
		return defaultValue
	}
	return input == "y" || input == "Y"
}

func addProvider() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	fileExisted := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fileExisted = false
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if !fileExisted {
		fmt.Println("Creating new config file...")
	}

	fmt.Println("\n=== Add OpenAI-Compatible Provider ===")

	providerKey := promptString("Provider key (e.g., ollama, custom)", "custom")
	displayName := promptString("Display name", "Custom Provider")
	baseURL := promptString("Base URL (e.g., http://localhost:11434/v1)", "http://localhost:11434/v1")
	apiKey := promptString("API key (optional)", "")

	provider := Provider{
		NPM:     "@ai-sdk/openai-compatible",
		Name:    displayName,
		Options: map[string]interface{}{"baseURL": baseURL},
		Models:  make(map[string]Model),
	}

	if apiKey != "" {
		provider.Options["apiKey"] = apiKey
	}

	if promptBool("Add custom headers?", false) {
		headers := make(map[string]string)
		for {
			headerName := promptString("Header name (leave blank to finish)", "")
			if headerName == "" {
				break
			}
			headerValue := promptString("Header value", "")
			if headerValue != "" {
				headers[headerName] = headerValue
			}
			if !promptBool("Add another header?", false) {
				break
			}
		}
		if len(headers) > 0 {
			provider.Options["headers"] = headers
		}
	}

	config.Provider[providerKey] = provider

	fmt.Println("\n=== Add Models ===")
	for {
		modelID := promptString("Model ID (e.g., qwen3-coder)", "")
		if modelID == "" {
			break
		}

		modelName := promptString("Display name", modelID)
		model := Model{Name: modelName}

		if promptBool("Configure token limits?", false) {
			contextLimit := promptString("Context limit (tokens, e.g., 128000)", "")
			outputLimit := promptString("Output limit (tokens, e.g., 65536)", "")

			if contextLimit != "" || outputLimit != "" {
				limit := &ModelLimit{}
				if contextLimit != "" {
					fmt.Sscanf(contextLimit, "%d", &limit.Context)
				}
				if outputLimit != "" {
					fmt.Sscanf(outputLimit, "%d", &limit.Output)
				}
				model.Limit = limit
			}
		}

		provider.Models[modelID] = model

		if !promptBool("Add another model?", false) {
			break
		}
	}

	if len(provider.Models) > 0 && (config.Model == "" || promptBool("Set as default model?", false)) {
		config.Model = fmt.Sprintf("%s/%s", providerKey, getFirstModelID(provider.Models))
	}

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("\nConfiguration saved to: %s\n", configPath)
	fmt.Printf("Added provider: %s with %d model(s)\n", displayName, len(provider.Models))
	if config.Model != "" {
		fmt.Printf("Default model: %s\n", config.Model)
	}
	return nil
}

func getFirstModelID(models map[string]Model) string {
	for id := range models {
		return id
	}
	return ""
}

func listProviders() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if len(config.Provider) == 0 {
		fmt.Println("No providers configured")
		return nil
	}

	fmt.Println("\n=== Configured Providers ===")
	for key, provider := range config.Provider {
		fmt.Printf("\nProvider: %s (%s)\n", provider.Name, key)
		fmt.Printf("  Base URL: %v\n", provider.Options["baseURL"])

		if headers, ok := provider.Options["headers"].(map[string]interface{}); ok && len(headers) > 0 {
			fmt.Println("  Custom headers:")
			for k, v := range headers {
				fmt.Printf("    %s: %v\n", k, v)
			}
		}

		if len(provider.Models) > 0 {
			fmt.Println("  Models:")
			for modelID, model := range provider.Models {
				fmt.Printf("    - %s (%s)", model.Name, modelID)
				if model.Limit != nil {
					if model.Limit.Context > 0 {
						fmt.Printf(" [context: %d]", model.Limit.Context)
					}
					if model.Limit.Output > 0 {
						fmt.Printf(" [output: %d]", model.Limit.Output)
					}
				}
				fmt.Println()
			}
		} else {
			fmt.Println("  Models: None")
		}
	}

	if config.Model != "" {
		fmt.Printf("\nDefault model: %s\n", config.Model)
	}
	if config.SmallModel != "" {
		fmt.Printf("Small model: %s\n", config.SmallModel)
	}
	if len(config.EnabledProviders) > 0 {
		fmt.Printf("Enabled providers: %v\n", config.EnabledProviders)
	}
	if len(config.DisabledProviders) > 0 {
		fmt.Printf("Disabled providers: %v\n", config.DisabledProviders)
	}
	return nil
}

func deleteProvider() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if len(config.Provider) == 0 {
		fmt.Println("No providers to delete")
		return nil
	}

	fmt.Println("\n=== Delete Provider ===")
	fmt.Println("Available providers:")
	keys := make([]string, 0, len(config.Provider))
	for key, provider := range config.Provider {
		fmt.Printf("  %s (%s)\n", key, provider.Name)
		keys = append(keys, key)
	}

	keyToDelete := promptString("Enter provider key to delete", "")
	if keyToDelete == "" {
		fmt.Println("Cancelled")
		return nil
	}

	if _, exists := config.Provider[keyToDelete]; !exists {
		fmt.Printf("Provider '%s' not found\n", keyToDelete)
		return nil
	}

	providerName := config.Provider[keyToDelete].Name
	delete(config.Provider, keyToDelete)

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Deleted provider: %s\n", providerName)
	return nil
}

func setDefaultModel() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if len(config.Provider) == 0 {
		fmt.Println("No providers configured. Use 'add' command first.")
		return nil
	}

	fmt.Println("\n=== Set Default Model ===")
	fmt.Println("Available models:")

	models := []string{}
	for providerKey, provider := range config.Provider {
		for modelID := range provider.Models {
			modelRef := fmt.Sprintf("%s/%s", providerKey, modelID)
			models = append(models, modelRef)
			fmt.Printf("  - %s (%s)\n", modelRef, provider.Models[modelID].Name)
		}
	}

	selectedModel := promptString("Enter model (provider/model)", "")
	if selectedModel == "" {
		fmt.Println("Cancelled")
		return nil
	}

	config.Model = selectedModel

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Default model set to: %s\n", selectedModel)
	return nil
}

func showHelp() {
	fmt.Println("OpenCode Configuration Wizard")
	fmt.Println("\nUsage:")
	fmt.Println("  opencode-config-wizard [command]")
	fmt.Println("\nCommands:")
	fmt.Println("  add          Add a new OpenAI-compatible provider")
	fmt.Println("  list         List all configured providers")
	fmt.Println("  delete       Delete a provider")
	fmt.Println("  set-default  Set default model")
	fmt.Println("  help         Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  opencode-config-wizard add")
	fmt.Println("  opencode-config-wizard list")
	fmt.Println("  opencode-config-wizard set-default")
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "add":
		if err := addProvider(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := listProviders(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		if err := deleteProvider(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "set-default":
		if err := setDefaultModel(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		showHelp()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
}
