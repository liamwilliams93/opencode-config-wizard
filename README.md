# OpenCode Config Wizard

A CLI tool to easily configure OpenCode with custom OpenAI-compatible API providers.

## Installation

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

### Delete a provider
```bash
./opencode-config-wizard delete
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
| `add` | Add a new OpenAI-compatible provider |
| `list` | List all configured providers and settings |
| `delete` | Delete a provider |
| `set-default` | Set the default model |
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
