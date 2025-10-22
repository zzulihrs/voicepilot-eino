package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	apiKey  = "sk-e6ffb92c2b365f6e1f6bb8ff1c2fe840077f15b500f2410a4564b2090df4d03e"
	baseURL = "https://openai.qiniu.com/v1"
)

func testChatAPI() error {
	log.Println("Testing Chat API...")

	reqBody := map[string]interface{}{
		"model": "deepseek/deepseek-v3.1-terminus",
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个有帮助的助手"},
			{"role": "user", "content": "你好，请介绍一下你自己"},
		},
		"max_tokens":  100,
		"temperature": 0.7,
		"stream":      false,
	}

	reqBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	log.Printf("Chat API Status: %d", resp.StatusCode)
	log.Printf("Chat API Response: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Chat API 返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return fmt.Errorf("响应中没有 choices")
	}

	log.Println("✅ Chat API 测试成功!")
	return nil
}

func testTTSAPI() error {
	log.Println("\nTesting TTS API...")

	reqBody := map[string]interface{}{
		"audio": map[string]interface{}{
			"voice_type":  "qiniu_zh_female_wwxkjx",
			"encoding":    "mp3",
			"speed_ratio": 1.0,
		},
		"request": map[string]string{
			"text": "你好，这是一个测试",
		},
	}

	reqBytes, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/voice/tts", bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	log.Printf("TTS API Status: %d", resp.StatusCode)
	log.Printf("TTS API Response: %s", string(respBody)[:200])

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TTS API 返回错误状态码: %d", resp.StatusCode)
	}

	log.Println("✅ TTS API 测试成功!")
	return nil
}

func main() {
	log.Println("开始测试七牛云 API...")
	log.Println("API Key:", apiKey[:20]+"...")

	// 测试 Chat API
	if err := testChatAPI(); err != nil {
		log.Printf("❌ Chat API 测试失败: %v\n", err)
	}

	// 测试 TTS API
	if err := testTTSAPI(); err != nil {
		log.Printf("❌ TTS API 测试失败: %v\n", err)
	}

	log.Println("\n测试完成!")
}
