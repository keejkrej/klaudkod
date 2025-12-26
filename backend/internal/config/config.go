package config

import (
	"os"
	"strconv"
)

type Config struct {
	LLMBaseURL       string
	LLMAPIKey        string
	LLMModel         string
	ServerPort       string
	ToolsEnabled     bool
	PermissionMode   string
	CommandTimeout   int
	WorkingDirectory string
}

func Load() *Config {
	return &Config{
		LLMBaseURL:       getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMAPIKey:        getEnv("LLM_API_KEY", ""),
		LLMModel:         getEnv("LLM_MODEL", "gpt-4"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		ToolsEnabled:     getEnvBool("TOOLS_ENABLED", true),
		PermissionMode:   getEnv("PERMISSION_MODE", "auto"),
		CommandTimeout:   getEnvInt("COMMAND_TIMEOUT", 120),
		WorkingDirectory: getEnv("WORKING_DIR", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}