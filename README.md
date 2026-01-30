# OpenCode Config Wizard

A CLI tool to easily configure OpenCode with custom OpenAI-compatible API providers.

## Installation

### Install with Go

```bash
go install github.com/liamwilliams93/opencode-config-wizard@latest
```

This will install the binary to `GOBIN` (or `GOPATH/bin`) as `opencode-config-wizard`. You can then run it from anywhere:

```bash
opencode-config-wizard help
```

### Build from source

**Windows:**
```bash
build.bat
```

**Linux/macOS:**
```bash
./build.sh
```

Or manually:
```bash
go build -o opencode-config-wizard .
```

**Note:** The config file will be created automatically when you run `add` command for the first time.

**Command name:** If you installed via `go install`, the command is `opencode-config-wizard`. If you built from source, use `opencode-config-wizard` (or `opencode-config-wizard.exe` on Windows).

## Usage

### List configured providers
```bash
./opencode-config-wizard list
```

### Add a new provider
```bash
./opencode-config-wizard add
```

Example with Ollama:
```
=== Add OpenAI-Compatible Provider ===
Provider key (e.g., ollama, custom) [custom]: ollama
Display name [Custom Provider]: Ollama
Base URL (e.g., http://localhost:11434/v1) [http://localhost:11434/v1]: http://localhost:11434/v1
API key (optional):
Add custom headers? [n] (y/n): n

=== Add Models ===
Model ID (e.g., qwen3-coder): qwen3-coder
Display name [qwen3-coder]: Qwen 3 Coder
Configure token limits? [n] (y/n): y
Context limit (tokens, e.g., 128000): 128000
Output limit (tokens, e.g., 65536): 65536
Add another model? [n] (y/n): n
Set as default model? [n] (y/n): y

Configuration saved to: C:\Users\liamw\AppData\Roaming\opencode\opencode.json
Added provider: Ollama with 1 model(s)
Default model: ollama/qwen3-coder
```

### Set default model
```bash
./opencode-config-wizard set-default
```

### Add model to existing provider
```bash
./opencode-config-wizard add-model
```

Example:
```
=== Add Model to Existing Provider ===
Available providers:
  1. test (test) - 1 model(s)
Enter provider number or key: 1

Adding model to provider: test (test)
Model ID (e.g., qwen3-coder): llama3
Display name [llama3]: Llama 3 70B
Configure token limits? [n] (y/n): n
Set as default model? [n] (y/n): n

Model 'Llama 3 70B' added to provider 'test'
```

### Delete a provider
```bash
./opencode-config-wizard delete
```

### Delete a model
```bash
./opencode-config-wizard delete-model
```

### Add an MCP server
```bash
./opencode-config-wizard add-mcp
```

Example with a local MCP server:
```
=== Add MCP Server ===
Server name (e.g., my-mcp): context7
Server type:
  1. Local (runs a command)
  2. Remote (connects to a URL)
Select type (1 or 2) [1]: 2

=== Remote MCP Server ===
Server URL (e.g., https://mcp.example.com/mcp): https://mcp.context7.com/mcp
Add custom headers? [n] (y/n): y
Header name (leave blank to finish): CONTEXT7_API_KEY
Header value: {env:CONTEXT7_API_KEY}
Add another header? [n] (y/n): n
Configure OAuth? [n] (y/n): n
Enable server on startup? [y] (y/n): y
Set custom timeout? [n] (y/n): n

Configuration saved to: C:\Users\liamw\.config\opencode\opencode.json
Added MCP server: context7 (type: remote)
Status: enabled
```

Example with a local MCP server:
```
=== Add MCP Server ===
Server name (e.g., my-mcp): mcp_everything
Server type:
  1. Local (runs a command)
  2. Remote (connects to a URL)
Select type (1 or 2) [1]: 1

=== Local MCP Server ===
Command (e.g., npx, bun) [npx]: npx
Arguments (e.g., -y @modelcontextprotocol/server-everything): -y @modelcontextprotocol/server-everything
Add another argument? [n] (y/n): n
Add environment variables? [n] (y/n): n
Enable server on startup? [y] (y/n): y
Set custom timeout? [n] (y/n): n

Configuration saved to: C:\Users\liamw\.config\opencode\opencode.json
Added MCP server: mcp_everything (type: local)
Status: enabled
```

### List MCP servers
```bash
./opencode-config-wizard list-mcp
```

### Delete an MCP server
```bash
./opencode-config-wizard delete-mcp
```

Example:
```
=== Delete Model ===
Available providers:
  1. test (test) - 2 model(s)
Enter provider number or key: 1

Provider: test (test)
Available models:
  1. qwen3-coder (qwen3-coder)
  2. testmodel (testmodel)
Enter model number or ID: 2

Are you sure you want to delete model 'testmodel' from provider 'test'? y
Deleted model: testmodel
```

## Config Location

Configuration is stored at:
- **All platforms**: `~/.config/opencode/opencode.json`
  - Windows: `C:\Users\<username>\.config\opencode\opencode.json`
  - Linux/macOS: `~/.config/opencode/opencode.json`

## Features

- **Multiple providers**: Configure multiple OpenAI-compatible providers
- **Custom headers**: Add custom HTTP headers for authentication or other purposes
- **Token limits**: Configure context and output token limits per model
- **Default model**: Set a default model for quick access
- **Provider management**: List, add, and delete providers easily
- **MCP servers**: Add, list, and delete local and remote MCP servers

## Example Config

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "ollama/qwen3-coder",
  "provider": {
    "ollama": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Ollama (local)",
      "options": {
        "baseURL": "http://localhost:11434/v1"
      },
      "models": {
        "qwen3-coder": {
          "name": "Qwen 3 Coder",
          "limit": {
            "context": 128000,
            "output": 65536
          }
        }
      }
    },
    "custom-provider": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "My Custom Provider",
      "options": {
        "baseURL": "https://api.example.com/v1",
        "apiKey": "{env:MY_API_KEY}",
        "headers": {
          "X-Custom-Header": "value"
        }
      },
      "models": {
        "my-model": {
          "name": "My Model Display Name"
        }
      }
    }
  }
}
```

## Commands

| Command | Description |
|---------|-------------|
| Provider Commands | |
| `add` | Add a new OpenAI-compatible provider |
| `add-model` | Add a model to an existing provider |
| `list` | List all configured providers and settings |
| `delete` | Delete a provider |
| `delete-model` | Delete a model from a provider |
| `set-default` | Set default model |
| MCP Server Commands | |
| `add-mcp` | Add a new MCP server (local or remote) |
| `list-mcp` | List all configured MCP servers |
| `delete-mcp` | Delete an MCP server |
| Other | |
| `help` | Show help message |

## Advanced Configuration

### Using Environment Variables
You can use `{env:VARIABLE_NAME}` syntax in the config:
```json
{
  "options": {
    "apiKey": "{env:OPENAI_API_KEY}"
  }
}
```

### Custom Headers
Add custom headers for authentication or other purposes:
```json
{
  "options": {
    "headers": {
      "Authorization": "Bearer token",
      "X-Custom-Header": "value"
    }
  }
}
```

### Token Limits
Configure context and output limits per model:
```json
{
  "models": {
    "model-id": {
      "name": "Model Display Name",
      "limit": {
        "context": 128000,
        "output": 65536
      }
    }
  }
}
```

## Documentation

For more information about OpenCode configuration, visit:
- https://opencode.ai/docs/config/
- https://opencode.ai/docs/providers/
- https://opencode.ai/docs/mcp-servers/
