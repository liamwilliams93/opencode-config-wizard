package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func showMainMenu() {
	fmt.Println("\nOpenCode Configuration Wizard")
	fmt.Println()
	fmt.Println("1. Provider Commands")
	fmt.Println("2. MCP Server Commands")
	fmt.Println("0. Exit")
}

func showProviderMenu() {
	fmt.Println("\nProvider Commands")
	fmt.Println()
	fmt.Println("1. List all configured providers")
	fmt.Println("2. Add a new OpenAI-compatible provider")
	fmt.Println("3. Add a model to an existing provider")
	fmt.Println("4. Delete a provider")
	fmt.Println("5. Delete a model from a provider")
	fmt.Println("6. Set default model")
	fmt.Println("0. Back to main menu")
}

func showMCPMenu() {
	fmt.Println("\nMCP Server Commands")
	fmt.Println()
	fmt.Println("1. List all configured MCP servers")
	fmt.Println("2. Add a new MCP server")
	fmt.Println("3. Delete an MCP server")
	fmt.Println("0. Back to main menu")
}

func getMenuChoice(maxOption int) int {
	fmt.Print("\nEnter choice: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if input == "" {
		return -1
	}

	choice, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}

	if choice < 0 || choice > maxOption {
		return -1
	}

	return choice
}

func executeWithErrorHandling(fn func() error) {
	fmt.Println()
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
	}
}

func runProviderMenu() {
	for {
		showProviderMenu()
		choice := getMenuChoice(6)

		switch choice {
		case 0:
			return
		case 1:
			executeWithErrorHandling(listProviders)
		case 2:
			executeWithErrorHandling(addProvider)
		case 3:
			executeWithErrorHandling(addModel)
		case 4:
			executeWithErrorHandling(deleteProvider)
		case 5:
			executeWithErrorHandling(deleteModel)
		case 6:
			executeWithErrorHandling(setDefaultModel)
		default:
			fmt.Println("\nInvalid choice, please try again")
		}
	}
}

func runMCPMenu() {
	for {
		showMCPMenu()
		choice := getMenuChoice(3)

		switch choice {
		case 0:
			return
		case 1:
			executeWithErrorHandling(listMCPServers)
		case 2:
			executeWithErrorHandling(addMCPServer)
		case 3:
			executeWithErrorHandling(deleteMCPServer)
		default:
			fmt.Println("\nInvalid choice, please try again")
		}
	}
}

func main() {
	fmt.Println("OpenCode Configuration Wizard")

	for {
		showMainMenu()
		choice := getMenuChoice(2)

		switch choice {
		case 0:
			fmt.Println("\nGoodbye!")
			return
		case 1:
			runProviderMenu()
		case 2:
			runMCPMenu()
		default:
			fmt.Println("\nInvalid choice, please try again")
		}
	}
}
