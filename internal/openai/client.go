package openai

import (
	"context"
	"os"

	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Client wraps the OpenAI client with our configuration
type Client struct {
	client openai.Client
	model  string
}

// NewClient creates a new OpenAI client with the given API key
func NewClient(cfg *config.Config) *Client {
	client := openai.NewClient(
		option.WithAPIKey(cfg.OpenAI.APIKey),
	)

	return &Client{
		client: client,
		model:  cfg.OpenAI.Model,
	}
}

// ChatCompletion handles chat completion requests
func (c *Client) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	// Create chat completion request
	chatCompletion, err := c.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    c.model,
		},
	)
	if err != nil {
		return nil, err
	}

	return chatCompletion, nil
}

// GetAPIKey returns the OpenAI API key from environment variable
func GetAPIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}
