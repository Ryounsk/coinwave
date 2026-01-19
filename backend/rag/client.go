package rag

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type VolcClient struct {
	client *resty.Client
}

// Internal structures for Volcengine API
type EmbeddingInputItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type EmbeddingRequest struct {
	Model string               `json:"model"`
	Input []EmbeddingInputItem `json:"input"`
}

type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingResponse struct {
	Data  json.RawMessage `json:"data"` // Can be Object or Array
	Error *APIError       `json:"error,omitempty"`
}

type ChatContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ChatMessageInput struct {
	Role    string            `json:"role"`
	Content []ChatContentItem `json:"content"`
}

type ChatRequest struct {
	Model string             `json:"model"`
	Input []ChatMessageInput `json:"input"`
}

type ChatResponse struct {
	Output []struct {
		Type    string `json:"type"` // "message" or "reasoning"
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error *APIError `json:"error,omitempty"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Public Message struct used by Service
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewVolcClient() *VolcClient {
	return &VolcClient{
		client: resty.New().
			SetHeader("Authorization", "Bearer "+VolcAuthToken).
			SetHeader("Content-Type", "application/json").
			SetTimeout(120 * time.Second), // Increased timeout for RAG
	}
}

func (c *VolcClient) GetEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Convert simple texts to multimodal input format
	inputs := make([]EmbeddingInputItem, len(texts))
	for i, text := range texts {
		inputs[i] = EmbeddingInputItem{
			Type: "text",
			Text: text,
		}
	}

	reqBody := EmbeddingRequest{
		Model: VolcModelEmbedding,
		Input: inputs,
	}

	var respBody EmbeddingResponse
	resp, err := c.client.R().
		SetBody(reqBody).
		SetResult(&respBody).
		Post(VolcEmbeddingEndpoint)

	if err != nil {
		log.Printf("[VolcClient] Network Error: %v", err)
		return nil, err
	}
	if resp.IsError() {
		log.Printf("[VolcClient] API Error Status: %s, Body: %s", resp.Status(), resp.String())
		return nil, fmt.Errorf("api error: %s", resp.String())
	}
	if respBody.Error != nil {
		log.Printf("[VolcClient] API Business Error: %s", respBody.Error.Message)
		return nil, fmt.Errorf("api error: %s", respBody.Error.Message)
	}

	// Handle Data parsing (Array or Object)
	var dataItems []EmbeddingData

	// Try Array first
	if err := json.Unmarshal(respBody.Data, &dataItems); err != nil {
		// Try Object
		var singleItem EmbeddingData
		if err := json.Unmarshal(respBody.Data, &singleItem); err == nil {
			dataItems = []EmbeddingData{singleItem}
		} else {
			return nil, fmt.Errorf("failed to parse embedding data: %v", err)
		}
	}

	embeddings := make([][]float32, len(texts))
	for i, item := range dataItems {
		// If index is present, use it; otherwise assume order is preserved
		idx := i
		if item.Index != 0 {
			idx = item.Index
		}
		if idx < len(embeddings) {
			embeddings[idx] = item.Embedding
		}
	}

	return embeddings, nil
}

func (c *VolcClient) Chat(messages []Message) (string, error) {
	// Convert simple messages to multimodal input format
	chatInputs := make([]ChatMessageInput, len(messages))
	for i, msg := range messages {
		chatInputs[i] = ChatMessageInput{
			Role: msg.Role,
			Content: []ChatContentItem{
				{
					Type: "input_text", // Note: API expects "input_text" here
					Text: msg.Content,
				},
			},
		}
	}

	reqBody := ChatRequest{
		Model: VolcModelLLM,
		Input: chatInputs,
	}

	var respBody ChatResponse
	resp, err := c.client.R().
		SetBody(reqBody).
		SetResult(&respBody).
		Post(VolcResponseEndpoint)

	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("api error: %s", resp.String())
	}
	if respBody.Error != nil {
		return "", fmt.Errorf("api error: %s", respBody.Error.Message)
	}

	// Parse output to find the assistant message
	for _, item := range respBody.Output {
		if item.Type == "message" && item.Role == "assistant" {
			if len(item.Content) > 0 {
				return item.Content[0].Text, nil
			}
		}
	}

	return "", errors.New("no assistant message found in response")
}
