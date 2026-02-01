package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
		if !promptBool(fmt.Sprintf("Server '%s' already exists. Overwrite?", serverName), false) {
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
	if mcpServer.Enabled == nil || *mcpServer.Enabled {
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

		status := "disabled"
		if server.Enabled == nil || *server.Enabled {
			status = "enabled"
		}
		fmt.Printf("  Status: %s\n", status)

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
	i := 1
	for name, server := range config.MCP {
		enabledStr := "disabled"
		if server.Enabled == nil || *server.Enabled {
			enabledStr = "enabled"
		}
		fmt.Printf("  %d. %s (%s) - %s\n", i, name, server.Type, enabledStr)
		keys = append(keys, name)
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

	nameToDelete := keys[choice-1]

	if !promptBool(fmt.Sprintf("Are you sure you want to delete MCP server '%s'?", nameToDelete), false) {
		fmt.Println("Cancelled")
		return nil
	}

	delete(config.MCP, nameToDelete)

	if err := saveConfig(config, configPath); err != nil {
		return err
	}

	fmt.Printf("Deleted MCP server: %s\n", nameToDelete)
	return nil
}
