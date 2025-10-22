package main

import (
	"context"
	"log"
	"os"

	"github.com/deca/voicepilot-eino/internal/config"
	"github.com/deca/voicepilot-eino/internal/qiniu"
)

func main() {
	log.Println("=== Dual-Strategy ASR Test ===")

	// Load configuration
	log.Println("Loading configuration...")
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Check environment variables
	log.Println("\n=== Environment Check ===")
	apiKey := os.Getenv("QINIU_API_KEY")
	accessKey := os.Getenv("QINIU_ACCESS_KEY")
	secretKey := os.Getenv("QINIU_SECRET_KEY")
	bucket := os.Getenv("QINIU_BUCKET")
	domain := os.Getenv("QINIU_DOMAIN")

	log.Printf("QINIU_API_KEY: %s", maskString(apiKey))
	log.Printf("QINIU_ACCESS_KEY: %s", maskString(accessKey))
	log.Printf("QINIU_SECRET_KEY: %s", maskString(secretKey))
	log.Printf("QINIU_BUCKET: %s", bucket)
	log.Printf("QINIU_DOMAIN: %s", domain)

	if accessKey == "" || secretKey == "" {
		log.Println("\n⚠️  Object storage credentials not configured")
		log.Println("    ASR will fall back to WebSocket method only")
	} else {
		log.Println("\n✅ Object storage credentials configured")
		log.Println("   ASR will try storage-based HTTP API first")
	}

	// Create test audio using TTS first
	log.Println("\n=== Step 1: Generate Test Audio ===")
	client := qiniu.NewClient()
	testText := "你好，这是一个语音识别测试"

	audioURL, err := client.TTS(context.Background(), testText)
	if err != nil {
		log.Fatalf("❌ TTS failed: %v", err)
	}

	log.Printf("✅ TTS generated audio: %s", audioURL)

	// Extract audio file path from URL
	audioPath := audioURL[1:] // Remove leading /
	log.Printf("Audio file path: %s", audioPath)

	// Test ASR (will try both strategies automatically)
	log.Println("\n=== Step 2: Testing ASR (Dual Strategy) ===")
	log.Printf("Original text: '%s'", testText)

	recognizedText, err := client.ASR(context.Background(), audioPath)
	if err != nil {
		log.Printf("\n❌ ASR FAILED: %v", err)
		log.Println("\n=== Troubleshooting Guide ===")

		if accessKey == "" || secretKey == "" {
			log.Println("1. Configure Object Storage (Recommended):")
			log.Println("   export QINIU_ACCESS_KEY='your_access_key'")
			log.Println("   export QINIU_SECRET_KEY='your_secret_key'")
			log.Println("   export QINIU_BUCKET='your_bucket_name'")
			log.Println("   export QINIU_DOMAIN='your_bucket_domain.com'")
			log.Println("")
		}

		log.Println("2. OR contact Qiniu Cloud support to enable WebSocket ASR permissions")
		log.Println("")
		log.Println("3. Use text input as alternative:")
		log.Println("   POST /api/process-text with {\"text\": \"your query\"}")

		os.Exit(1)
	}

	log.Println("\n=== Test Results ===")
	log.Printf("✅ Original text:   '%s'", testText)
	log.Printf("✅ Recognized text: '%s'", recognizedText)

	// Compare results
	if recognizedText != "" {
		log.Println("\n✅✅✅ ASR TEST PASSED!")
		log.Println("ASR functionality is working correctly")
	} else {
		log.Println("\n❌ ASR TEST FAILED: Empty recognition result")
		os.Exit(1)
	}
}

func maskString(s string) string {
	if s == "" {
		return "<not set>"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
