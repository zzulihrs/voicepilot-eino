package qiniu

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/deca/voicepilot-eino/internal/config"
)

// Client is the Qiniu Cloud API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Qiniu Cloud API client
func NewClient() *Client {
	return &Client{
		apiKey:  config.AppConfig.QiniuAPIKey,
		baseURL: config.AppConfig.QiniuBaseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ASR performs speech-to-text conversion
func (c *Client) ASR(ctx context.Context, audioPath string) (string, error) {
	log.Printf("Starting ASR for audio file: %s", audioPath)

	// Read audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	// Encode to base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	// Construct data URL
	dataURL := fmt.Sprintf("data:audio/%s;base64,%s", config.AppConfig.ASRFormat, audioBase64)

	// Build request
	reqBody := map[string]interface{}{
		"model": config.AppConfig.ASRModel,
		"audio": map[string]string{
			"format": config.AppConfig.ASRFormat,
			"url":    dataURL,
		},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/voice/asr"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("ASR API error response: %s", string(respBody))
		return "", fmt.Errorf("ASR API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("ASR completed successfully: %s", result.Text)
	return result.Text, nil
}

// TTS performs text-to-speech conversion
func (c *Client) TTS(ctx context.Context, text string) (string, error) {
	log.Printf("Starting TTS for text: %s", text)

	// Build request
	reqBody := map[string]interface{}{
		"audio": map[string]interface{}{
			"voice_type":  config.AppConfig.TTSVoiceType,
			"encoding":    config.AppConfig.TTSEncoding,
			"speed_ratio": config.AppConfig.TTSSpeedRatio,
		},
		"request": map[string]string{
			"text": text,
		},
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/voice/tts"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("TTS API error response: %s", string(respBody))
		return "", fmt.Errorf("TTS API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Handle different response formats
	var audioURL string
	var audioData string

	// Check for direct URL
	if url, ok := result["url"].(string); ok {
		audioURL = url
	} else if data, ok := result["audio"].(string); ok {
		// Check if it's base64 data
		audioData = data
	} else if audioMap, ok := result["audio"].(map[string]interface{}); ok {
		if data, ok := audioMap["data"].(string); ok {
			audioData = data
		}
	}

	// If we have base64 data, decode and save it
	if audioData != "" {
		decodedData, err := base64.StdEncoding.DecodeString(audioData)
		if err != nil {
			return "", fmt.Errorf("failed to decode audio data: %w", err)
		}

		// Save to static audio path
		filename := fmt.Sprintf("tts_%d.%s", time.Now().Unix(), config.AppConfig.TTSEncoding)
		filepath := filepath.Join(config.AppConfig.StaticAudioPath, filename)

		if err := os.WriteFile(filepath, decodedData, 0644); err != nil {
			return "", fmt.Errorf("failed to save audio file: %w", err)
		}

		audioURL = fmt.Sprintf("/static/audio/%s", filename)
	}

	if audioURL == "" {
		return "", fmt.Errorf("no audio URL or data in response")
	}

	log.Printf("TTS completed successfully: %s", audioURL)
	return audioURL, nil
}

// ChatCompletion performs LLM chat completion
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	log.Printf("Starting chat completion with %d messages", len(messages))

	// Build request
	reqBody := map[string]interface{}{
		"model":       config.AppConfig.LLMModel,
		"messages":    messages,
		"max_tokens":  config.AppConfig.LLMMaxTokens,
		"temperature": config.AppConfig.LLMTemperature,
		"stream":      false,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Chat API error response: %s", string(respBody))
		return "", fmt.Errorf("Chat API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := result.Choices[0].Message.Content
	log.Printf("Chat completion successful: %s", content)
	return content, nil
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
