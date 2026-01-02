package backend

import (
	"time"
)

// Source represents a document source added to a notebook
type Source struct {
	ID          string                 `json:"id"`
	NotebookID  string                 `json:"notebook_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "file", "url", "text", "youtube"
	URL         string                 `json:"url,omitempty"`
	Content     string                 `json:"content,omitempty"`
	FileName    string                 `json:"file_name,omitempty"`
	FileSize    int64                  `json:"file_size,omitempty"`
	ChunkCount  int                    `json:"chunk_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Note represents a note generated from sources
type Note struct {
	ID          string                 `json:"id"`
	NotebookID  string                 `json:"notebook_id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"` // "summary", "faq", "study_guide", "outline", "custom"
	SourceIDs   []string               `json:"source_ids"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Notebook represents a collection of sources and notes
type Notebook struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID         string                 `json:"id"`
	SessionID  string                 `json:"session_id"`
	Role       string                 `json:"role"` // "user", "assistant", "system"
	Content    string                 `json:"content"`
	Sources    []string               `json:"sources,omitempty"` // Source IDs referenced
	CreatedAt  time.Time              `json:"created_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ChatSession represents a chat session within a notebook
type ChatSession struct {
	ID           string                 `json:"id"`
	NotebookID   string                 `json:"notebook_id"`
	Title        string                 `json:"title"`
	Messages     []ChatMessage          `json:"messages"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Podcast represents an audio podcast generated from sources
type Podcast struct {
	ID          string                 `json:"id"`
	NotebookID  string                 `json:"notebook_id"`
	Title       string                 `json:"title"`
	Script      string                 `json:"script"`
	AudioURL    string                 `json:"audio_url,omitempty"`
	Duration    int                    `json:"duration,omitempty"` // in seconds
	Voice       string                 `json:"voice"`
	Status      string                 `json:"status"` // "pending", "generating", "completed", "error"
	SourceIDs   []string               `json:"source_ids"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TransformationRequest represents a request to generate a note
type TransformationRequest struct {
	Type       string   `json:"type"`       // "summary", "faq", "study_guide", "outline", "podcast", "custom"
	Prompt     string   `json:"prompt"`     // Custom prompt for "custom" type
	SourceIDs  []string `json:"source_ids"` // Specific sources to use, empty = all
	Length     string   `json:"length"`     // "short", "medium", "long"
	Format     string   `json:"format"`     // "markdown", "bullet_points", "paragraphs"
}

// TransformationResponse represents the response from a transformation
type TransformationResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Sources   []SourceSummary        `json:"sources"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SourceSummary is a lightweight source reference
type SourceSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message   string                 `json:"message"`
	SessionID string                 `json:"session_id,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message     string                 `json:"message"`
	Sources     []SourceSummary        `json:"sources"`
	SessionID   string                 `json:"session_id"`
	MessageID   string                 `json:"message_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp int64             `json:"timestamp"`
	Services  map[string]string `json:"services"`
}
