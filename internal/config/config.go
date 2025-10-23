package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port string

	// Qiniu Cloud API configuration
	QiniuAPIKey   string
	QiniuBaseURL  string
	TTSVoiceType  string
	TTSEncoding   string
	TTSSpeedRatio float64

	// ASR configuration
	ASRModel  string
	ASRFormat string

	// LLM configuration
	LLMModel       string
	LLMMaxTokens   int
	LLMTemperature float64

	// Audio storage
	StaticAudioPath string
	TempAudioPath   string

	// Session and context management
	SessionStoragePath  string
	SessionMaxHistory   int
	SessionExpiryHours  int

	// Security
	EnableSafeMode bool
	MaxAudioSize   int64 // in bytes
}

var AppConfig *Config

// Load loads configuration from environment variables
func Load() error {
	// Load .env file if exists (ignore error if not found)
	_ = godotenv.Load()

	AppConfig = &Config{
		Port:               getEnv("PORT", "8080"),
		QiniuAPIKey:        getEnv("QINIU_API_KEY", ""),
		QiniuBaseURL:       getEnv("QINIU_BASE_URL", "https://openai.qiniu.com/v1"),
		TTSVoiceType:       getEnv("TTS_VOICE_TYPE", "qiniu_zh_female_wwxkjx"),
		TTSEncoding:        getEnv("TTS_ENCODING", "mp3"),
		TTSSpeedRatio:      getEnvFloat("TTS_SPEED_RATIO", 1.0),
		ASRModel:           getEnv("ASR_MODEL", "asr"),
		ASRFormat:          getEnv("ASR_FORMAT", "wav"),
		LLMModel:           getEnv("LLM_MODEL", "deepseek/deepseek-v3.1-terminus"),
		LLMMaxTokens:       getEnvInt("LLM_MAX_TOKENS", 2000),
		LLMTemperature:     getEnvFloat("LLM_TEMPERATURE", 0.7),
		StaticAudioPath:    getEnv("STATIC_AUDIO_PATH", "./static/audio"),
		TempAudioPath:      getEnv("TEMP_AUDIO_PATH", "./temp"),
		SessionStoragePath: getEnv("SESSION_STORAGE_PATH", "./data/sessions"),
		SessionMaxHistory:  getEnvInt("SESSION_MAX_HISTORY", 50),
		SessionExpiryHours: getEnvInt("SESSION_EXPIRY_HOURS", 72),
		EnableSafeMode:     getEnvBool("ENABLE_SAFE_MODE", true),
		MaxAudioSize:       getEnvInt64("MAX_AUDIO_SIZE", 10*1024*1024), // 10MB default
	}

	// Validate required configuration
	if AppConfig.QiniuAPIKey == "" {
		return fmt.Errorf("QINIU_API_KEY is required")
	}

	// Ensure directories exist
	if err := os.MkdirAll(AppConfig.StaticAudioPath, 0755); err != nil {
		return fmt.Errorf("failed to create static audio directory: %w", err)
	}
	if err := os.MkdirAll(AppConfig.TempAudioPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp audio directory: %w", err)
	}
	if err := os.MkdirAll(AppConfig.SessionStoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create session storage directory: %w", err)
	}

	return nil
}

// Helper functions for getting environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
