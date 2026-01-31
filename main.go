package main

import (
	"fmt"
	"os"
)

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
