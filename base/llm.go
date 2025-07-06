package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// represents a request to the OpenAI API
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

// represents a message in the OpenAI API
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// represents a response from the OpenAI API
type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error,omitempty"`
}

// represents a choice in the OpenAI response
type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

// represents an error from the OpenAI API
type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (c *LLMClient) GetReponse(ctx context.Context, message string, model Model) (string, error) {
	if !model.IsValid() {
		return "", fmt.Errorf("invalid model: %s", model.String())
	}

	switch model {
	case Llama2, Llama3:
		return c.getResponseFromOllama(ctx, message, model)
	case GPT4o, GPT35Turbo:
		return c.getResponseFromOpenAI(ctx, message, model)
	default:
		return "", fmt.Errorf("unsupported model: %s", model.String())
	}
}

// this generates a response using the default model
func (c *LLMClient) GetResponseWithDefaultModel(ctx context.Context, message string) (string, error) {
	return c.GetReponse(ctx, message, c.config.DefaultModel)
}

// this handles requests to the Ollama API
func (c *LLMClient) getResponseFromOllama(ctx context.Context, message string, model Model) (string, error) {
	req := ollamaRequest{
		ModelName: model.String(),
		Prompt:    message,
		Stream:    false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.OllamaURL+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API returned status %d", httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ollama response: %w", err)
	}

	var resp ollamaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal ollama response: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("ollama API error: %s", resp.Error)
	}

	return resp.Response, nil
}

func (m Model) IsValid() bool {
	return m >= Llama2 && m <= Llama3
}
