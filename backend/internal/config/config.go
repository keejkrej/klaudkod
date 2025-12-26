package config

import "os"

type Config struct {
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
	ServerPort string
}

func Load() *Config {
	return &Config{
		LLMBaseURL: getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMAPIKey:  getEnv("LLM_API_KEY", ""),
		LLMModel:   getEnv("LLM_MODEL", "gpt-4"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
