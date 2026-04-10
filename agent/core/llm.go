package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const defaultLLMRequestTimeout = 2 * time.Minute

type streamParser struct {
	tools     map[int]*streamToolCall
	content   strings.Builder
	reasoning strings.Builder
}

type streamToolCall struct {
	id       string
	typeName string
	name     string
	args     strings.Builder
}

func newStreamParser() *streamParser {
	return &streamParser{
		tools: make(map[int]*streamToolCall),
	}
}

func requestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, defaultLLMRequestTimeout)
}

type LLMClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewLLMClient(baseURL, apiKey, model string) *LLMClient {
	return &LLMClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 0, // streaming may be long-lived
		},
	}
}

func parseStreamLine(parser *streamParser, raw string) (tokens []string) {
	line := strings.TrimSpace(raw)
	if line == "" || line == "[DONE]" {
		return nil
	}
	if !strings.HasPrefix(line, "data:") {
		return nil
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))

	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil
	}

	choices, ok := chunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil
	}

	ch0, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil
	}

	delta, ok := ch0["delta"].(map[string]interface{})
	if !ok {
		return nil
	}

	for _, key := range []string{"thinking", "thought", "reasoning", "reasoning_content"} {
		if v, ok := delta[key].(string); ok && v != "" {
			tokens = append(tokens, v)
			parser.reasoning.WriteString(v)
		}
	}

	if content, ok := delta["content"].(string); ok && content != "" {
		tokens = append(tokens, content)
		parser.content.WriteString(content)
	}

	if toolCalls, ok := delta["tool_calls"].([]interface{}); ok {
		for _, tc := range toolCalls {
			toolMap, ok := tc.(map[string]interface{})
			if !ok {
				continue
			}

			indexVal, ok := toolMap["index"].(float64)
			if !ok {
				continue
			}
			index := int(indexVal)

			funcData, ok := toolMap["function"].(map[string]interface{})
			if !ok {
				continue
			}

			if _, ok := parser.tools[index]; !ok {
				parser.tools[index] = &streamToolCall{}
			}

			t := parser.tools[index]
			if id, ok := toolMap["id"].(string); ok {
				t.id = id
			}
			if typ, ok := toolMap["type"].(string); ok {
				t.typeName = typ
			}

			if name, ok := funcData["name"].(string); ok {
				t.name = name
			}

			if args, ok := funcData["arguments"].(string); ok {
				t.args.WriteString(args)
			}
		}
	}

	return tokens
}

func streamUsageEstimate(content string, toolCalls []ToolCall) int {
	total := utf8.RuneCountInString(content)
	for _, toolCall := range toolCalls {
		total += utf8.RuneCountInString(toolCall.Function.Arguments)
		total += utf8.RuneCountInString(toolCall.Function.Name)
	}
	if total <= 0 {
		return 1
	}
	estimate := total / 3
	if estimate <= 0 {
		return 1
	}
	return estimate
}

func estimateChatTokens(messages []ChatMessage, tools []FunctionTool) int {
	total := 0

	for _, message := range messages {
		total += utf8.RuneCountInString(string(message.Role))
		total += utf8.RuneCountInString(message.Content)
		total += utf8.RuneCountInString(message.ToolCallID)
		for _, toolCall := range message.ToolCalls {
			total += utf8.RuneCountInString(toolCall.Id)
			total += utf8.RuneCountInString(toolCall.Type)
			total += utf8.RuneCountInString(toolCall.Function.Name)
			total += utf8.RuneCountInString(toolCall.Function.Arguments)
		}
	}

	for _, tool := range tools {
		total += utf8.RuneCountInString(tool.Type)
		total += utf8.RuneCountInString(tool.Function.Name)
		total += utf8.RuneCountInString(tool.Function.Description)

		paramBytes, err := json.Marshal(tool.Function.Parameters)
		if err == nil {
			total += len(paramBytes)
		}
	}

	if total <= 0 {
		return 1
	}

	estimate := total / 3
	if estimate <= 0 {
		return 1
	}
	return estimate
}

func (c *LLMClient) streamResultFromParser(parser *streamParser, rawContent string) LLMResult {
	result := LLMResult{}
	content := strings.TrimSpace(parser.content.String())
	reasoning := strings.TrimSpace(parser.reasoning.String())
	if content == "" {
		content = strings.TrimSpace(rawContent)
	}

	var keys []int
	for key := range parser.tools {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	toolCalls := make([]ToolCall, 0, len(keys))
	for _, key := range keys {
		tool := parser.tools[key]
		toolCalls = append(toolCalls, ToolCall{
			Id:   tool.id,
			Type: tool.typeName,
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      tool.name,
				Arguments: tool.args.String(),
			},
		})
	}

	result.Choices = []LLMChoice{{
		Message: LLMMessage{
			Content:          content,
			ReasoningContent: reasoning,
			ToolCalls:        toolCalls,
		},
	}}
	result.Usage.TotalTokens = streamUsageEstimate(content+reasoning, toolCalls)
	return result
}

func (c *LLMClient) Models(ctx context.Context) []string {
	url := strings.TrimRight(c.BaseURL, "/") + "/models"

	ctx, cancel := requestContext(ctx)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(bodyBytes, &modelsResp); err != nil {
		return nil
	}

	result := []string{}
	for _, model := range modelsResp.Data {
		result = append(result, model.ID)
	}
	return result
}

func (c *LLMClient) Chat(ctx context.Context, messages []ChatMessage, tools []FunctionTool, temperature float64) (LLMResult, error) {
	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	req := ChatRequest{Model: c.Model, Messages: messages, Tools: tools, Temperature: temperature}
	body, err := json.Marshal(req)
	if err != nil {
		return LLMResult{}, err
	}
	ctx, cancel := requestContext(ctx)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return LLMResult{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return LLMResult{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return LLMResult{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return LLMResult{}, err
	}

	result := LLMResult{}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return LLMResult{}, fmt.Errorf("decode response: %w; body: %s", err, string(bodyBytes))
	}
	if len(result.Choices) == 0 {
		return LLMResult{}, errors.New("no choices returned")
	}

	return result, nil
}

func (c *LLMClient) StreamChat(ctx context.Context, messages []ChatMessage, tools []FunctionTool, temperature float64, onToken func(string) error) (LLMResult, error) {
	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	req := ChatRequest{Model: c.Model, Messages: messages, Tools: tools, Stream: true, Temperature: temperature}
	promptTokens := estimateChatTokens(messages, tools)
	body, err := json.Marshal(req)
	if err != nil {
		return LLMResult{}, err
	}
	ctx, cancel := requestContext(ctx)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return LLMResult{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return LLMResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return LLMResult{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	parser := newStreamParser()
	var streamedContent strings.Builder
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				result := c.streamResultFromParser(parser, streamedContent.String())
				result.Usage.PromptTokens = promptTokens
				result.Usage.TotalTokens = promptTokens + result.Usage.TotalTokens
				return result, nil
			}
			return LLMResult{}, err
		}
		tokens := parseStreamLine(parser, line)
		for _, t := range tokens {
			if t == "" {
				continue
			}
			streamedContent.WriteString(t)
			if onToken != nil {
				if err := onToken(t); err != nil {
					return LLMResult{}, err
				}
			}
		}
	}
}
