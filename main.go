package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Schema            string               `json:"$schema"`
	Provider          map[string]Provider  `json:"provider"`
	Model             string               `json:"model,omitempty"`
	SmallModel        string               `json:"small_model,omitempty"`
	EnabledProviders  []string             `json:"enabled_providers,omitempty"`
	DisabledProviders []string             `json:"disabled_providers,omitempty"`
	MCP               map[string]MCPServer `json:"mcp,omitempty"`
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

type MCPServer struct {
	Type        string                 `json:"type"`
	Command     []string               `json:"command,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	OAuth       map[string]interface{} `json:"oauth,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Timeout     *int                   `json:"timeout,omitempty"`
}

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

func promptString(prompt string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

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

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

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

func deleteModel() error {
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

	fmt.Println("\n=== Delete Model ===")
	fmt.Println("Available providers:")

	providers := []string{}
	i := 1
	for key, provider := range config.Provider {
		fmt.Printf("  %d. %s (%s) - %d model(s)\n", i, key, provider.Name, len(provider.Models))
		providers = append(providers, key)
		i++
	}

	selection := promptString("Enter provider number or key", "")
	var providerKey string

	if selection == "" {
		fmt.Println("Cancelled")
		return nil
	}

	num := 0
	if _, err := fmt.Sscanf(selection, "%d", &num); err == nil && num > 0 && num <= len(providers) {
		providerKey = providers[num-1]
	} else {
		providerKey = selection
	}

	if _, exists := config.Provider[providerKey]; !exists {
		fmt.Printf("Provider '%s' not found\n", providerKey)
		return nil
	}

	provider := config.Provider[providerKey]
	if len(provider.Models) == 0 {
		fmt.Printf("Provider '%s' has no models to delete\n", provider.Name)
		return nil
	}

	fmt.Printf("\nProvider: %s (%s)\n", provider.Name, providerKey)
	fmt.Println("Available models:")

	modelKeys := []string{}
	j := 1
	for modelID, model := range provider.Models {
		fmt.Printf("  %d. %s (%s)\n", j, model.Name, modelID)
		modelKeys = append(modelKeys, modelID)
		j++
	}

	modelSelection := promptString("Enter model number or ID", "")
	var modelID string

	if modelSelection == "" {
		fmt.Println("Cancelled")
		return nil
	}

	if _, err := fmt.Sscanf(modelSelection, "%d", &num); err == nil && num > 0 && num <= len(modelKeys) {
		modelID = modelKeys[num-1]
	} else {
		modelID = modelSelection
	}

	if _, exists := provider.Models[modelID]; !exists {
		fmt.Printf("Model '%s' not found\n", modelID)
		return nil
	}

	modelName := provider.Models[modelID].Name

	fmt.Printf("\nAre you sure you want to delete model '%s' from provider '%s'? ", modelName, provider.Name)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	confirm := strings.TrimSpace(scanner.Text())
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Cancelled")
		return nil
	}

	delete(provider.Models, modelID)

	if config.Model == fmt.Sprintf("%s/%s", providerKey, modelID) {
		fmt.Printf("Warning: This was the default model. Default model cleared.\n")
		config.Model = ""
	}

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Deleted model: %s\n", modelName)
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

func addMCPServer() error {
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

	fmt.Println("\n=== Add MCP Server ===")

	serverName := promptString("Server name (e.g., my-mcp)", "")
	if serverName == "" {
		fmt.Println("Cancelled")
		return nil
	}

	if _, exists := config.MCP[serverName]; exists {
		fmt.Printf("Server '%s' already exists. Overwrite? ", serverName)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		confirm := strings.TrimSpace(scanner.Text())
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	fmt.Println("Server type:")
	fmt.Println("  1. Local (runs a command)")
	fmt.Println("  2. Remote (connects to a URL)")

	typeSelection := promptString("Select type (1 or 2)", "1")
	serverType := "local"
	if typeSelection == "2" {
		serverType = "remote"
	}

	mcpServer := MCPServer{
		Type: serverType,
	}

	if serverType == "local" {
		fmt.Println("\n=== Local MCP Server ===")

		command := promptString("Command (e.g., npx, bun)", "npx")
		args := promptString("Arguments (e.g., -y @modelcontextprotocol/server-everything)", "")

		cmdArray := []string{command}
		if args != "" {
			cmdArray = append(cmdArray, strings.Fields(args)...)
		}

		for {
			if !promptBool("Add another argument?", false) {
				break
			}
			arg := promptString("Additional argument", "")
			if arg != "" {
				cmdArray = append(cmdArray, arg)
			}
		}

		mcpServer.Command = cmdArray

		if promptBool("Add environment variables?", false) {
			envVars := make(map[string]string)
			for {
				envName := promptString("Environment variable name (leave blank to finish)", "")
				if envName == "" {
					break
				}
				envValue := promptString("Environment variable value", "")
				if envValue != "" {
					envVars[envName] = envValue
				}
				if !promptBool("Add another environment variable?", false) {
					break
				}
			}
			if len(envVars) > 0 {
				mcpServer.Environment = envVars
			}
		}
	} else {
		fmt.Println("\n=== Remote MCP Server ===")
		url := promptString("Server URL (e.g., https://mcp.example.com/mcp)", "")
		if url == "" {
			fmt.Println("URL is required for remote servers")
			return nil
		}
		mcpServer.URL = url

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
				mcpServer.Headers = headers
			}
		}

		if promptBool("Configure OAuth?", false) {
			oauthConfig := make(map[string]interface{})
			clientId := promptString("Client ID (leave blank for dynamic registration)", "")
			if clientId != "" {
				oauthConfig["clientId"] = clientId
			}
			clientSecret := promptString("Client Secret (optional)", "")
			if clientSecret != "" {
				oauthConfig["clientSecret"] = clientSecret
			}
			scope := promptString("OAuth scopes (optional)", "")
			if scope != "" {
				oauthConfig["scope"] = scope
			}
			if len(oauthConfig) > 0 {
				mcpServer.OAuth = oauthConfig
			}
		}
	}

	enabled := promptBool("Enable server on startup?", true)
	if !enabled {
		mcpServer.Enabled = &enabled
	}

	if promptBool("Set custom timeout?", false) {
		timeoutStr := promptString("Timeout in milliseconds (default: 5000)", "")
		if timeoutStr != "" {
			var timeout int
			fmt.Sscanf(timeoutStr, "%d", &timeout)
			mcpServer.Timeout = &timeout
		}
	}

	config.MCP[serverName] = mcpServer

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("\nConfiguration saved to: %s\n", configPath)
	fmt.Printf("Added MCP server: %s (type: %s)\n", serverName, serverType)
	if mcpServer.Enabled != nil && *mcpServer.Enabled {
		fmt.Println("Status: enabled")
	} else {
		fmt.Println("Status: disabled")
	}
	return nil
}

func listMCPServers() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if len(config.MCP) == 0 {
		fmt.Println("No MCP servers configured")
		return nil
	}

	fmt.Println("\n=== Configured MCP Servers ===")
	for name, server := range config.MCP {
		fmt.Printf("\nServer: %s\n", name)
		fmt.Printf("  Type: %s\n", server.Type)

		if server.Enabled != nil {
			status := "disabled"
			if *server.Enabled {
				status = "enabled"
			}
			fmt.Printf("  Status: %s\n", status)
		}

		if server.Type == "local" {
			if len(server.Command) > 0 {
				fmt.Printf("  Command: %v\n", server.Command)
			}
			if len(server.Environment) > 0 {
				fmt.Println("  Environment variables:")
				for k, v := range server.Environment {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
		} else {
			if server.URL != "" {
				fmt.Printf("  URL: %s\n", server.URL)
			}
			if len(server.Headers) > 0 {
				fmt.Println("  Headers:")
				for k, v := range server.Headers {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			if len(server.OAuth) > 0 {
				fmt.Println("  OAuth configured")
			}
		}

		if server.Timeout != nil {
			fmt.Printf("  Timeout: %d ms\n", *server.Timeout)
		}
	}
	return nil
}

func deleteMCPServer() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if len(config.MCP) == 0 {
		fmt.Println("No MCP servers to delete")
		return nil
	}

	fmt.Println("\n=== Delete MCP Server ===")
	fmt.Println("Available servers:")
	keys := make([]string, 0, len(config.MCP))
	for name, server := range config.MCP {
		enabledStr := "disabled"
		if server.Enabled != nil && *server.Enabled {
			enabledStr = "enabled"
		}
		fmt.Printf("  %s (%s) - %s\n", name, server.Type, enabledStr)
		keys = append(keys, name)
	}

	nameToDelete := promptString("Enter server name to delete", "")
	if nameToDelete == "" {
		fmt.Println("Cancelled")
		return nil
	}

	if _, exists := config.MCP[nameToDelete]; !exists {
		fmt.Printf("Server '%s' not found\n", nameToDelete)
		return nil
	}

	delete(config.MCP, nameToDelete)

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Deleted MCP server: %s\n", nameToDelete)
	return nil
}

func addModel() error {
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

	fmt.Println("\n=== Add Model to Existing Provider ===")
	fmt.Println("Available providers:")

	providers := []string{}
	i := 1
	for key, provider := range config.Provider {
		fmt.Printf("  %d. %s (%s) - %d model(s)\n", i, key, provider.Name, len(provider.Models))
		providers = append(providers, key)
		i++
	}

	selection := promptString("Enter provider number or key", "")
	var providerKey string

	if selection == "" {
		fmt.Println("Cancelled")
		return nil
	}

	num := 0
	if _, err := fmt.Sscanf(selection, "%d", &num); err == nil && num > 0 && num <= len(providers) {
		providerKey = providers[num-1]
	} else {
		providerKey = selection
	}

	if _, exists := config.Provider[providerKey]; !exists {
		fmt.Printf("Provider '%s' not found\n", providerKey)
		return nil
	}

	provider := config.Provider[providerKey]
	fmt.Printf("\nAdding model to provider: %s (%s)\n", provider.Name, providerKey)

	modelID := promptString("Model ID (e.g., qwen3-coder)", "")
	if modelID == "" {
		fmt.Println("Cancelled")
		return nil
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

	if _, exists := provider.Models[modelID]; exists {
		fmt.Printf("\nWarning: Model '%s' already exists. Overwrite? ", modelID)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		confirm := strings.TrimSpace(scanner.Text())
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	provider.Models[modelID] = model

	if promptBool("Set as default model?", false) {
		config.Model = fmt.Sprintf("%s/%s", providerKey, modelID)
	}

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("\nModel '%s' added to provider '%s'\n", modelName, provider.Name)
	if config.Model == fmt.Sprintf("%s/%s", providerKey, modelID) {
		fmt.Printf("Default model: %s\n", config.Model)
	}
	return nil
}

func showHelp() {
	fmt.Println("OpenCode Configuration Wizard")
	fmt.Println("\nUsage:")
	fmt.Println("  opencode-config-wizard [command]")
	fmt.Println("\nProvider Commands:")
	fmt.Println("  add          Add a new OpenAI-compatible provider")
	fmt.Println("  add-model    Add a model to an existing provider")
	fmt.Println("  list         List all configured providers")
	fmt.Println("  delete       Delete a provider")
	fmt.Println("  delete-model Delete a model from a provider")
	fmt.Println("  set-default  Set default model")
	fmt.Println("\nMCP Server Commands:")
	fmt.Println("  add-mcp      Add a new MCP server")
	fmt.Println("  list-mcp     List all configured MCP servers")
	fmt.Println("  delete-mcp   Delete an MCP server")
	fmt.Println("\nOther:")
	fmt.Println("  help         Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  opencode-config-wizard add")
	fmt.Println("  opencode-config-wizard add-mcp")
	fmt.Println("  opencode-config-wizard list")
	fmt.Println("  opencode-config-wizard list-mcp")
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
	case "add-model":
		if err := addModel(); err != nil {
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
	case "delete-model":
		if err := deleteModel(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "set-default":
		if err := setDefaultModel(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add-mcp":
		if err := addMCPServer(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list-mcp":
		if err := listMCPServers(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "delete-mcp":
		if err := deleteMCPServer(); err != nil {
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
