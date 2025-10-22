package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env
	originalKey := os.Getenv("QINIU_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("QINIU_API_KEY", originalKey)
		}
	}()

	// Test with missing API key
	os.Unsetenv("QINIU_API_KEY")
	err := Load()
	if err == nil {
		t.Error("Expected error when QINIU_API_KEY is missing")
	}

	// Test with valid API key
	os.Setenv("QINIU_API_KEY", "test-api-key")
	err = Load()
	if err != nil {
		t.Errorf("Expected no error with valid API key, got: %v", err)
	}

	if AppConfig == nil {
		t.Fatal("AppConfig should not be nil after successful load")
	}

	if AppConfig.QiniuAPIKey != "test-api-key" {
		t.Errorf("Expected QiniuAPIKey to be 'test-api-key', got: %s", AppConfig.QiniuAPIKey)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test-value")
	defer os.Unsetenv("TEST_KEY")

	result := getEnv("TEST_KEY", "default")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got: %s", result)
	}

	result = getEnv("NON_EXISTENT_KEY", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got: %s", result)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := getEnvInt("TEST_INT", 0)
	if result != 42 {
		t.Errorf("Expected 42, got: %d", result)
	}

	result = getEnvInt("NON_EXISTENT_INT", 10)
	if result != 10 {
		t.Errorf("Expected 10, got: %d", result)
	}

	os.Setenv("INVALID_INT", "not-a-number")
	result = getEnvInt("INVALID_INT", 5)
	if result != 5 {
		t.Errorf("Expected default value 5 for invalid int, got: %d", result)
	}
	os.Unsetenv("INVALID_INT")
}

func TestGetEnvFloat(t *testing.T) {
	os.Setenv("TEST_FLOAT", "3.14")
	defer os.Unsetenv("TEST_FLOAT")

	result := getEnvFloat("TEST_FLOAT", 0.0)
	if result != 3.14 {
		t.Errorf("Expected 3.14, got: %f", result)
	}

	result = getEnvFloat("NON_EXISTENT_FLOAT", 1.0)
	if result != 1.0 {
		t.Errorf("Expected 1.0, got: %f", result)
	}
}

func TestGetEnvBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	result := getEnvBool("TEST_BOOL", false)
	if result != true {
		t.Errorf("Expected true, got: %v", result)
	}

	result = getEnvBool("NON_EXISTENT_BOOL", false)
	if result != false {
		t.Errorf("Expected false, got: %v", result)
	}

	os.Setenv("TEST_BOOL", "false")
	result = getEnvBool("TEST_BOOL", true)
	if result != false {
		t.Errorf("Expected false, got: %v", result)
	}
}

func TestConfigDefaults(t *testing.T) {
	os.Setenv("QINIU_API_KEY", "test-key")
	defer os.Unsetenv("QINIU_API_KEY")

	// Clear all other env vars
	os.Unsetenv("PORT")
	os.Unsetenv("QINIU_BASE_URL")
	os.Unsetenv("TTS_VOICE_TYPE")

	err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if AppConfig.Port != "8080" {
		t.Errorf("Expected default port 8080, got: %s", AppConfig.Port)
	}

	if AppConfig.QiniuBaseURL != "https://openai.qiniu.com/v1" {
		t.Errorf("Expected default Qiniu base URL, got: %s", AppConfig.QiniuBaseURL)
	}

	if AppConfig.TTSVoiceType != "qiniu_zh_female_wwxkjx" {
		t.Errorf("Expected default TTS voice type, got: %s", AppConfig.TTSVoiceType)
	}

	if AppConfig.EnableSafeMode != true {
		t.Error("Expected safe mode to be enabled by default")
	}
}
