package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

	if len(provider.Models) > 0 && promptBool("Set as default model?", false) {
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
	i := 1
	for key, provider := range config.Provider {
		fmt.Printf("  %d. %s (%s)\n", i, key, provider.Name)
		keys = append(keys, key)
		i++
	}

	choice := getMenuChoice(len(keys))
	if choice == -1 {
		fmt.Println("Invalid choice")
		return nil
	}
	if choice == 0 {
		fmt.Println("Cancelled")
		return nil
	}

	keyToDelete := keys[choice-1]

	providerName := config.Provider[keyToDelete].Name

	if !promptBool(fmt.Sprintf("Are you sure you want to delete provider '%s'?", providerName), false) {
		fmt.Println("Cancelled")
		return nil
	}

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
	fmt.Println("Available models:")

	models := []string{}
	i := 1
	for providerKey, provider := range config.Provider {
		for modelID := range provider.Models {
			modelRef := fmt.Sprintf("%s/%s", providerKey, modelID)
			models = append(models, modelRef)
			fmt.Printf("  %d. %s (%s)\n", i, modelRef, provider.Models[modelID].Name)
			i++
		}
	}

	if len(models) == 0 {
		fmt.Println("No models configured")
		return nil
	}

	choice := getMenuChoice(len(models))
	if choice == -1 {
		fmt.Println("Invalid choice")
		return nil
	}
	if choice == 0 {
		fmt.Println("Cancelled")
		return nil
	}

	selectedModel := models[choice-1]

	parts := strings.Split(selectedModel, "/")
	if len(parts) != 2 {
		fmt.Printf("Invalid model reference: %s\n", selectedModel)
		return nil
	}

	providerKey := parts[0]
	modelID := parts[1]

	provider, exists := config.Provider[providerKey]
	if !exists {
		fmt.Printf("Provider '%s' not found\n", providerKey)
		return nil
	}

	model, exists := provider.Models[modelID]
	if !exists {
		fmt.Printf("Model '%s' not found\n", modelID)
		return nil
	}

	if !promptBool(fmt.Sprintf("\nAre you sure you want to delete model '%s' from provider '%s'?", model.Name, provider.Name), false) {
		fmt.Println("Cancelled")
		return nil
	}

	delete(provider.Models, modelID)

	if config.Model == selectedModel {
		fmt.Printf("Warning: This was the default model. Default model cleared.\n")
		config.Model = ""
	}

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Deleted model: %s\n", model.Name)
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
	i := 1
	for providerKey, provider := range config.Provider {
		for modelID := range provider.Models {
			modelRef := fmt.Sprintf("%s/%s", providerKey, modelID)
			models = append(models, modelRef)
			fmt.Printf("  %d. %s (%s)\n", i, modelRef, provider.Models[modelID].Name)
			i++
		}
	}

	choice := getMenuChoice(len(models))
	if choice == -1 {
		fmt.Println("Invalid choice")
		return nil
	}
	if choice == 0 {
		fmt.Println("Cancelled")
		return nil
	}

	selectedModel := models[choice-1]
	config.Model = selectedModel

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Default model set to: %s\n", selectedModel)
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
		if !promptBool(fmt.Sprintf("\nWarning: Model '%s' already exists. Overwrite?", modelID), false) {
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
