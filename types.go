package main

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
