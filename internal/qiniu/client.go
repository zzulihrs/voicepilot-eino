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
	"os/exec"
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

// detectAudioFormat detects audio format from file header
func detectAudioFormat(data []byte) string {
	if len(data) < 12 {
		return "wav" // default fallback
	}

	// WebM/Matroska: starts with 0x1A 0x45 0xDF 0xA3
	if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
		return "webm"
	}

	// WAV: starts with "RIFF" at offset 0 and "WAVE" at offset 8
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return "wav"
	}

	// MP3: starts with "ID3" or 0xFF 0xFB
	if len(data) >= 3 {
		if string(data[0:3]) == "ID3" {
			return "mp3"
		}
		if data[0] == 0xFF && (data[1]&0xE0) == 0xE0 {
			return "mp3"
		}
	}

	// Opus/Ogg: starts with "OggS"
	if len(data) >= 4 && string(data[0:4]) == "OggS" {
		return "opus"
	}

	return "wav" // default fallback
}

// convertToWav converts any audio format to WAV using ffmpeg or afconvert
func convertToWav(inputPath string) (string, error) {
	outputPath := inputPath + ".converted.wav"

	// Try ffmpeg first (more common across platforms)
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-ar", "16000", "-ac", "1", "-y", outputPath)
	if err := cmd.Run(); err != nil {
		log.Printf("ffmpeg conversion failed, trying afconvert: %v", err)

		// Fallback to afconvert (macOS)
		cmd = exec.Command("afconvert", "-f", "WAVE", "-d", "LEI16@16000", inputPath, outputPath)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("audio conversion failed (tried ffmpeg and afconvert): %w", err)
		}
	}

	log.Printf("Successfully converted audio to WAV: %s", outputPath)
	return outputPath, nil
}

// ASR performs speech-to-text conversion
func (c *Client) ASR(ctx context.Context, audioPath string) (string, error) {
	log.Printf("Starting ASR for audio file: %s", audioPath)

	// Strategy 1: Try HTTP REST API with storage upload (most reliable)
	result, err := c.ASRWithStorage(ctx, audioPath)
	if err == nil && result != "" {
		return result, nil
	}
	log.Printf("Storage-based ASR not available: %v", err)

	// Strategy 2: Try WebSocket ASR (implemented but may need permissions)
	result, err = c.WebSocketASR(ctx, audioPath)
	if err == nil && result != "" {
		return result, nil
	}
	log.Printf("WebSocket ASR not available: %v", err)

	// Both methods failed
	return "", fmt.Errorf("语音识别暂不可用。请配置七牛云对象存储(QINIU_ACCESS_KEY/SECRET_KEY)或使用文字输入")
}

// ASRWithStorage uploads audio to storage and uses HTTP REST API
func (c *Client) ASRWithStorage(ctx context.Context, audioPath string) (string, error) {
	// Upload to storage
	publicURL, err := c.UploadAudioToStorage(ctx, audioPath)
	if err != nil {
		return "", err
	}

	log.Printf("Using HTTP ASR with public URL: %s", publicURL)

	// Call HTTP REST API
	reqBody := map[string]interface{}{
		"model": "asr",
		"audio": map[string]interface{}{
			"format": "wav",
			"url":    publicURL,
		},
	}

	reqBytes, _ := json.Marshal(reqBody)
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/voice/asr", bytes.NewReader(reqBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ASR API error %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	// Extract text from response structure
	if data, ok := result["data"].(map[string]interface{}); ok {
		if resultData, ok := data["result"].(map[string]interface{}); ok {
			if text, ok := resultData["text"].(string); ok {
				log.Printf("ASR recognized: %s", text)
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response format")
}

// ASRWebSocket performs speech-to-text conversion using WebSocket (new implementation)
func (c *Client) ASRWebSocket(ctx context.Context, audioPath string) (string, error) {
	return c.WebSocketASR(ctx, audioPath)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	} else if data, ok := result["data"].(string); ok {
		// TTS response format: {"data": "base64..."}
		audioData = data
	} else if audioField, ok := result["audio"].(string); ok {
		// Alternative format
		audioData = audioField
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
		savePath := filepath.Join(config.AppConfig.StaticAudioPath, filename)

		if err := os.WriteFile(savePath, decodedData, 0644); err != nil {
			return "", fmt.Errorf("failed to save audio file: %w", err)
		}

		audioURL = fmt.Sprintf("/static/audio/%s", filename)
		log.Printf("TTS audio saved to: %s", audioURL)
	}

	if audioURL == "" {
		log.Printf("TTS response structure: %+v", result)
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
