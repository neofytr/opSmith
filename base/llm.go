package base

import (
	"net/http"
	"time"
)

type Model int

const (
	Llama2 Model = iota + 1
	GPT4o
	GPT35Turbo
	Llama3
)

const (
	_ollamaURL string = "http://localhost:11434/api/generate"
)

func (m Model) String() string {
	switch m {
	case Llama2:
		return "llama2"
	case GPT4o:
		return "gpt4o"
	case GPT35Turbo:
		return "gpt-3.5-turbo"
	case Llama3:
		return "llama3"
	default:
		return "unknown"
	}
}

// config holds the configuration for the LLM clients
type Config struct {
	DefaultModel Model
	OllamaURL    string
	OpenAIAPIKey string
	Timeout      time.Duration
}

// returns default configuration for an LLM client
func DefaultConfig() *Config {
	return &Config{
		DefaultModel: Llama2,
		OllamaURL:    _ollamaURL,
		Timeout:      120 * time.Second,
	}
}

// represents an LLM client
type LLMClient struct {
	config     *Config
	httpClient *http.Client
}

// get a new LLM client, possibly with default configuration
func NewLLMClient(config *Config) *LLMClient {
	if config == nil {
		config = DefaultConfig()
	}

	return &LLMClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type ollamaRequest struct {
	ModelName string `json:"model"`
	Prompt    string `json:"prompt"`
	Stream    bool   `json:"stream"`
}

type ollamaResponse struct {
	ModelName string `json:"model"`
	Done      bool   `json:"done"`
	Error     string `json:"error,omitempty"`
}

func (m Model) IsValid() bool {
	return m >= Llama2 && m <= Llama3
}
