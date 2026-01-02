package backend

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Server settings
	ServerHost string
	ServerPort string

	// LLM settings
	OpenAIAPIKey      string
	OpenAIBaseURL     string
	OpenAIModel       string
	EmbeddingModel    string
	OllamaBaseURL     string
	OllamaModel       string

	// Vector store settings
	VectorStoreType    string // "memory", "supabase", "pgvector", "redis", "sqlite"
	SupabaseURL        string
	SupabaseKey        string
	PostgreSQLURL      string
	RedisURL           string
	SQLitePath         string

	// Store settings (for checkpoints)
	StoreType          string // "memory", "sqlite", "postgres", "redis"
	StorePath          string

	// Application settings
	MaxSources         int
	MaxContextLength   int
	ChunkSize          int
	ChunkOverlap       int

	// Podcast generation
	EnablePodcast      bool
	PodcastVoice       string

	// Document conversion
	EnableMarkitdown   bool

	// LangSmith tracing (optional)
	LangChainAPIKey    string
	LangChainProject   string
}

// loadEnv loads .env file if it exists (ignoring errors if file not found)
func loadEnv() {
	// Try to load .env file from current directory
	_ = godotenv.Load()

	// Also try to load from .env.local for local overrides
	_ = godotenv.Load(".env.local")
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() Config {
	// Load .env file first (if exists)
	loadEnv()

	cfg := Config{
		ServerHost:       getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:    getEnv("OPENAI_BASE_URL", ""),
		OpenAIModel:      getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		EmbeddingModel:   getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		OllamaBaseURL:    getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:      getEnv("OLLAMA_MODEL", "llama3.2"),
		VectorStoreType:  getEnv("VECTOR_STORE_TYPE", "sqlite"),
		SupabaseURL:      getEnv("SUPABASE_URL", ""),
		SupabaseKey:      getEnv("SUPABASE_KEY", ""),
		PostgreSQLURL:    getEnv("POSTGRES_URL", ""),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		SQLitePath:       getEnv("SQLITE_PATH", "./data/vector.db"),
		StoreType:        getEnv("STORE_TYPE", "sqlite"),
		StorePath:        getEnv("STORE_PATH", "./data/checkpoints.db"),
		MaxSources:       getEnvInt("MAX_SOURCES", 5),
		MaxContextLength: getEnvInt("MAX_CONTEXT_LENGTH", 128000),
		ChunkSize:        getEnvInt("CHUNK_SIZE", 1000),
		ChunkOverlap:     getEnvInt("CHUNK_OVERLAP", 200),
		EnablePodcast:    getEnvBool("ENABLE_PODCAST", true),
		PodcastVoice:     getEnv("PODCAST_VOICE", "alloy"),
		EnableMarkitdown: getEnvBool("ENABLE_MARKITDOWN", true),
		LangChainAPIKey:  getEnv("LANGCHAIN_API_KEY", ""),
		LangChainProject: getEnv("LANGCHAIN_PROJECT", "open-notebook"),
	}

	// Auto-detect provider from base URL or model name
	if cfg.OpenAIBaseURL == "" && cfg.OpenAIModel != "" {
		if contains(cfg.OpenAIModel, "ollama") || contains(cfg.OpenAIModel, "llama") {
			cfg.OpenAIBaseURL = cfg.OllamaBaseURL
		}
	}

	return cfg
}

// ValidateConfig validates the configuration
func ValidateConfig(cfg Config) error {
	// Check if at least one LLM provider is configured
	hasOpenAI := cfg.OpenAIAPIKey != ""
	hasOllama := cfg.OpenAIBaseURL != "" && contains(cfg.OpenAIBaseURL, "11434")

	if !hasOpenAI && !hasOllama {
		return fmt.Errorf("either OPENAI_API_KEY or OLLAMA_BASE_URL must be set")
	}

	// Validate vector store configuration
	switch cfg.VectorStoreType {
	case "supabase":
		if cfg.SupabaseURL == "" || cfg.SupabaseKey == "" {
			return fmt.Errorf("SUPABASE_URL and SUPABASE_KEY required for supabase vector store")
		}
	case "pgvector", "postgres":
		if cfg.PostgreSQLURL == "" {
			return fmt.Errorf("POSTGRES_URL required for postgres vector store")
		}
	case "redis":
		// Redis URL has default
	case "sqlite":
		// SQLite path has default
	case "memory":
		// No validation needed
	default:
		return fmt.Errorf("unknown vector store type: %s", cfg.VectorStoreType)
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as a boolean or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetBaseURL returns the appropriate base URL for the LLM provider
func (c *Config) GetBaseURL() string {
	if c.OpenAIBaseURL != "" {
		return c.OpenAIBaseURL
	}
	return ""
}

// IsOllama returns true if using Ollama as the LLM provider
func (c *Config) IsOllama() bool {
	return c.OpenAIBaseURL != "" && contains(c.OpenAIBaseURL, "11434")
}

// SupportsFunctionCalling returns true if the configured model supports function calling
func (c *Config) SupportsFunctionCalling() bool {
	if c.IsOllama() {
		return true // Most Ollama models support tool calling now
	}
	// OpenAI models that support function calling
	supportingModels := []string{"gpt-4", "gpt-3.5-turbo"}
	for _, model := range supportingModels {
		if contains(c.OpenAIModel, model) {
			return true
		}
	}
	return contains(c.OpenAIModel, "gpt-4") || contains(c.OpenAIModel, "gpt-3.5-turbo")
}
