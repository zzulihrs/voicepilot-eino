package main

import (
	"context"
	"log"
	"os"

	"github.com/deca/voicepilot-eino/internal/config"
	"github.com/deca/voicepilot-eino/internal/qiniu"
)

func main() {
	log.Println("=== WebSocket ASR Test ===")

	// Load configuration
	log.Println("Loading configuration...")
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create test audio using TTS first
	log.Println("Step 1: Generate test audio using TTS...")

	client := qiniu.NewClient()
	testText := "你好，这是一个WebSocket语音识别测试"

	audioURL, err := client.TTS(context.Background(), testText)
	if err != nil {
		log.Fatalf("TTS failed: %v", err)
	}

	log.Printf("TTS generated audio URL: %s", audioURL)

	// Extract audio file path from URL (remove leading /)
	// audioURL format: /static/audio/tts_xxx.mp3
	audioPath := audioURL[1:] // Remove leading /

	log.Printf("Audio file path: %s", audioPath)

	// Test WebSocket ASR directly with MP3 (it will auto-convert)
	log.Println("Step 2: Testing WebSocket ASR...")
	log.Printf("Original text: %s", testText)

	recognizedText, err := client.ASR(context.Background(), audioPath)
	if err != nil {
		log.Fatalf("WebSocket ASR failed: %v", err)
	}

	log.Println("\n=== Test Results ===")
	log.Printf("✅ Original text: %s", testText)
	log.Printf("✅ Recognized text: %s", recognizedText)

	if recognizedText != "" {
		log.Println("\n✅ WebSocket ASR test PASSED!")
	} else {
		log.Println("\n❌ WebSocket ASR test FAILED: empty recognition result")
		os.Exit(1)
	}
}
