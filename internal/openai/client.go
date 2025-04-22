package openai

import (
	"context"
	"github.com/mcp-ecosystem/mcp-gateway/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
)

// Client wraps the OpenAI client with our configuration
type Client struct {
	client openai.Client
	model  string
}

// NewClient creates a new OpenAI client with the given API key
func NewClient(cfg *config.OpenAIConfig) *Client {
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)

	return &Client{
		client: client,
		model:  cfg.Model,
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

// ChatCompletionStream handles streaming chat completion requests
func (c *Client) ChatCompletionStream(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*ssestream.Stream[openai.ChatCompletionChunk], error) {
	// Create streaming chat completion request
	stream := c.client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    c.model,
		},
	)

	return stream, nil
}
