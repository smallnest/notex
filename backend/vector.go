package backend

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tmc/langchaingo/schema"
)

// VectorStore wraps different vector store implementations
type VectorStore struct {
	cfg  Config
	docs []schema.Document
	mu   sync.RWMutex
}

// VectorStats contains statistics about the vector store
type VectorStats struct {
	TotalDocuments int
	TotalVectors   int
	Dimension      int
}

// NewVectorStore creates a new vector store based on configuration
func NewVectorStore(cfg Config) (*VectorStore, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &VectorStore{
		cfg:  cfg,
		docs: make([]schema.Document, 0),
	}, nil
}

// IngestDocuments loads and indexes documents from file paths
func (vs *VectorStore) IngestDocuments(ctx context.Context, paths []string) error {
	for _, path := range paths {
		fmt.Printf("[VectorStore] Loading file: %s\n", path)

		content, err := vs.ExtractDocument(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to extract document %s: %w", path, err)
		}

		fmt.Printf("[VectorStore] File loaded, size: %d bytes\n", len(content))
		if err := vs.IngestText(ctx, filepath.Base(path), content); err != nil {
			return err
		}
	}

	return nil
}

// ExtractDocument reads and converts a document to text/markdown
func (vs *VectorStore) ExtractDocument(ctx context.Context, path string) (string, error) {
	// Check if file needs markitdown conversion
	ext := strings.ToLower(filepath.Ext(path))
	if vs.cfg.EnableMarkitdown && vs.needsMarkitdown(ext) {
		return vs.convertWithMarkitdown(path)
	}

	// Direct read for text files or when markitdown is disabled
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// IngestText ingests raw text content
func (vs *VectorStore) IngestText(ctx context.Context, sourceName, content string) error {
	// Split content into chunks
	chunks := vs.splitText(content, vs.cfg.ChunkSize, vs.cfg.ChunkOverlap)

	vs.mu.Lock()
	defer vs.mu.Unlock()

	// Create documents
	for i, chunk := range chunks {
		doc := schema.Document{
			PageContent: chunk,
			Metadata: map[string]any{
				"source": sourceName,
				"chunk":  i,
			},
		}
		vs.docs = append(vs.docs, doc)
	}

	fmt.Printf("[VectorStore] Ingested %d chunks from source '%s' (total docs: %d)\n", len(chunks), sourceName, len(vs.docs))
	return nil
}

// splitText splits text into chunks
func (vs *VectorStore) splitText(text string, chunkSize, chunkOverlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if chunkOverlap < 0 {
		chunkOverlap = 200
	}

	fmt.Printf("[VectorStore] Splitting text (len=%d, chunkSize=%d, overlap=%d)\n", len(text), chunkSize, chunkOverlap)

	var chunks []string

	// Check if text contains mostly CJK characters (Chinese, Japanese, Korean)
	runes := []rune(text)
	cjkCount := 0
	for _, r := range runes {
		if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified Ideographs
			cjkCount++
		}
	}
	cjkRatio := float64(cjkCount) / float64(len(runes))

	if cjkRatio > 0.3 {
		// For CJK text, split by character count (runes)
		fmt.Println("[VectorStore] Using CJK splitting (by character count)")
		for i := 0; i < len(runes); i += (chunkSize - chunkOverlap) {
			end := i + chunkSize
			if end > len(runes) {
				end = len(runes)
			}

			chunk := string(runes[i:end])
			chunks = append(chunks, chunk)

			if end >= len(runes) {
				break
			}
		}
	} else {
		// For Western text, split by words
		fmt.Println("[VectorStore] Using word-based splitting")
		words := strings.Fields(text)

		for i := 0; i < len(words); i += (chunkSize - chunkOverlap) {
			end := i + chunkSize
			if end > len(words) {
				end = len(words)
			}

			chunk := strings.Join(words[i:end], " ")
			chunks = append(chunks, chunk)

			if end >= len(words) {
				break
			}
		}
	}

	fmt.Printf("[VectorStore] Created %d chunks\n", len(chunks))
	return chunks
}

// SimilaritySearch performs a similarity search (simple keyword matching for now)
func (vs *VectorStore) SimilaritySearch(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	if numDocs <= 0 {
		numDocs = 5
	}

	vs.mu.RLock()
	defer vs.mu.RUnlock()

	fmt.Printf("[VectorStore] Searching for '%s' (total docs: %d)\n", query, len(vs.docs))

	if len(vs.docs) == 0 {
		fmt.Println("[VectorStore] No documents available for search")
		return []schema.Document{}, nil
	}

	// For Chinese and general text, use substring matching
	// Also extract individual words for English
	queryLower := strings.ToLower(query)
	queryRunes := []rune(queryLower)

	type docScore struct {
		doc   schema.Document
		score float64
	}

	scores := make([]docScore, 0, len(vs.docs))
	for _, doc := range vs.docs {
		content := strings.ToLower(doc.PageContent)
		score := 0.0

		// 1. Check if query appears as substring in content (good for Chinese)
		if strings.Contains(content, queryLower) {
			score += 10.0
		}

		// 2. For each character in query, check if it appears in content
		// This helps with partial matches
		matchCount := 0
		for _, r := range queryRunes {
			if strings.ContainsRune(content, r) {
				matchCount++
			}
		}
		if matchCount > 0 {
			charMatchRatio := float64(matchCount) / float64(len(queryRunes))
			score += charMatchRatio * 5.0
		}

		// 3. Word-based matching for English/Space-separated languages
		queryWords := strings.Fields(queryLower)
		for _, word := range queryWords {
			if len(word) > 2 && strings.Contains(content, word) {
				score += 2.0
			}
		}

		// 4. Check for common question keywords in Chinese
		questionKeywords := []string{"介绍", "什么", "啥", "内容", "文档", "说"}
		for _, keyword := range questionKeywords {
			if strings.Contains(queryLower, keyword) {
				// If query asks about the document, boost all documents
				score += 1.0
				break
			}
		}

		if score > 0 {
			scores = append(scores, docScore{doc: doc, score: score})
		}
	}

	fmt.Printf("[VectorStore] Found %d matching documents\n", len(scores))

	// Sort by score descending
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// If no matches found, return all documents (fallback)
	// This allows the LLM to use the full context
	if len(scores) == 0 {
		fmt.Println("[VectorStore] No matches found, returning all documents as fallback")
		result := make([]schema.Document, 0, min(numDocs, len(vs.docs)))
		for i := 0; i < len(result); i++ {
			result = append(result, vs.docs[i])
		}
		return result, nil
	}

	// Return top results
	result := make([]schema.Document, 0, numDocs)
	for i := 0; i < len(scores) && i < numDocs; i++ {
		result = append(result, scores[i].doc)
	}

	if len(result) > 0 {
		fmt.Printf("[VectorStore] Returning top %d results (best score: %.2f)\n", len(result), scores[0].score)
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Delete removes documents by source
func (vs *VectorStore) Delete(ctx context.Context, source string) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	filtered := make([]schema.Document, 0, len(vs.docs))
	for _, doc := range vs.docs {
		if docSource, ok := doc.Metadata["source"].(string); !ok || docSource != source {
			filtered = append(filtered, doc)
		}
	}
	vs.docs = filtered

	return nil
}

// GetStats returns statistics about the vector store
func (vs *VectorStore) GetStats(ctx context.Context) (VectorStats, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	stats := VectorStats{
		TotalDocuments: len(vs.docs),
		Dimension:      1536, // Default for OpenAI embeddings
	}

	if vs.cfg.IsOllama() {
		stats.Dimension = 768 // Common for Ollama models
	}

	return stats, nil
}

// needsMarkitdown checks if a file extension requires markitdown conversion
func (vs *VectorStore) needsMarkitdown(ext string) bool {
	markitdownExts := map[string]bool{
		".pdf":  true,
		".docx": true,
		".doc":  true,
		".pptx": true,
		".ppt":  true,
		".xlsx": true,
		".xls":  true,
	}
	return markitdownExts[ext]
}

// convertWithMarkitdown converts a document to Markdown using the markitdown CLI tool
func (vs *VectorStore) convertWithMarkitdown(filePath string) (string, error) {
	fmt.Printf("[VectorStore] Converting with markitdown: %s\n", filePath)

	// Create temporary output file
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("markitdown_%s.md", filepath.Base(filePath)))

	// Run markitdown command
	cmd := exec.Command("markitdown", filePath, "-o", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[VectorStore] markitdown error: %s\n", string(output))
		return "", fmt.Errorf("markitdown conversion failed: %w, output: %s", err, string(output))
	}

	// Read the converted markdown content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to read markitdown output: %w", err)
	}

	// Clean up temporary file
	os.Remove(tmpFile)

	fmt.Printf("[VectorStore] markitdown conversion successful, output size: %d bytes\n", len(content))
	return string(content), nil
}
