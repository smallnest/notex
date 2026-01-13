package backend

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/kataras/golog"
	"github.com/tmc/langchaingo/llms"
	ollamallm "github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

// Agent handles AI operations for generating notes and chat responses
type Agent struct {
	vectorStore *VectorStore
	llm         llms.Model
	cfg         Config
	provider    LLMProvider
}

// NewAgent creates a new agent
func NewAgent(cfg Config, vectorStore *VectorStore) (*Agent, error) {
	llm, err := createLLM(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	provider := NewGeminiClient(cfg.GoogleAPIKey, llm)

	return &Agent{
		vectorStore: vectorStore,
		llm:         llm,
		cfg:         cfg,
		provider:    provider,
	}, nil
}

// createLLM creates an LLM based on configuration
func createLLM(cfg Config) (llms.Model, error) {
	if cfg.IsOllama() {
		return ollamallm.New(
			ollamallm.WithModel(cfg.OllamaModel),
			ollamallm.WithServerURL(cfg.OllamaBaseURL),
		)
	}

	opts := []openai.Option{
		openai.WithToken(cfg.OpenAIAPIKey),
		openai.WithModel(cfg.OpenAIModel),
	}
	if cfg.OpenAIBaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.OpenAIBaseURL))
	}

	return openai.New(opts...)
}

// GenerateTransformation generates a note based on transformation type
func (a *Agent) GenerateTransformation(ctx context.Context, req *TransformationRequest, sources []Source) (*TransformationResponse, error) {
	// Build context from sources
	var sourceContext strings.Builder
	for i, src := range sources {
		sourceContext.WriteString(fmt.Sprintf("\n## Source %d: %s\n", i+1, src.Name))

		// Use MaxContextLength from config, or default to a safe large value if not set (or too small)
		limit := a.cfg.MaxContextLength
		if limit <= 0 {
			limit = 100000 // Default to 100k chars if config is invalid
		}

		if src.Content != "" {
			if len(src.Content) <= limit {
				sourceContext.WriteString(src.Content)
			} else {
				// Truncate content instead of replacing it entirely
				sourceContext.WriteString(src.Content[:limit])
				sourceContext.WriteString(fmt.Sprintf("\n... [Content truncated, total length: %d]", len(src.Content)))
			}
		} else {
			sourceContext.WriteString(fmt.Sprintf("[Source content: %s, type: %s]", src.Name, src.Type))
		}
		sourceContext.WriteString("\n")
	}

	// Build prompt using f-string format (no Go template reserved names issue)
	promptTemplate := getTransformationPrompt(req.Type)

	prompt := prompts.NewPromptTemplate(
		promptTemplate,
		[]string{"sources", "type", "length", "format", "prompt"},
	)
	prompt.TemplateFormat = prompts.TemplateFormatFString

	promptValue, err := prompt.Format(map[string]any{
		"sources": sourceContext.String(),
		"type":    req.Type,
		"length":  req.Length,
		"format":  req.Format,
		"prompt":  req.Prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// Generate response
	var response string
	var genErr error

	if req.Type == "ppt" {
		response, genErr = a.provider.GenerateTextWithModel(ctx, promptValue, "gemini-3-flash-preview")
	} else if req.Type == "insight" {
		// For insight type: first generate a summary, then call DeepInsight
		ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
		defer cancel()

		// Step 1: Generate summary
		summary, err := a.provider.GenerateFromSinglePrompt(ctx, a.llm, promptValue)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary: %w", err)
		}

		// Step 2: Call DeepInsight with the summary
		response, err = a.callDeepInsight(ctx, summary)
		if err != nil {
			return nil, fmt.Errorf("failed to generate deep insight: %w", err)
		}
	} else {
		ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
		defer cancel()
		response, genErr = a.provider.GenerateFromSinglePrompt(ctx, a.llm, promptValue)
	}

	if genErr != nil {
		return nil, fmt.Errorf("failed to generate response: %w", genErr)
	}

	// Build source summaries
	sourceSummaries := make([]SourceSummary, len(sources))
	for i, src := range sources {
		sourceSummaries[i] = SourceSummary{
			ID:   src.ID,
			Name: src.Name,
			Type: src.Type,
		}
	}

	return &TransformationResponse{
		Type:      req.Type,
		Content:   response,
		Sources:   sourceSummaries,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"length": req.Length,
			"format": req.Format,
		},
	}, nil
}

// Chat performs a chat query with RAG
func (a *Agent) Chat(ctx context.Context, notebookID, message string, history []ChatMessage) (*ChatResponse, error) {
	chatResult, err := a.buildChatPrompt(ctx, message, history)
	if err != nil {
		return nil, err
	}

	// Generate response
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	response, err := a.provider.GenerateFromSinglePrompt(ctx, a.llm, chatResult.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	return &ChatResponse{
		Message:   response,
		Sources:   chatResult.Sources,
		SessionID: notebookID,
		Metadata: map[string]interface{}{
			"docs_retrieved": chatResult.DocsRetrieved,
		},
	}, nil
}

// ChatStream 使用 RAG 进行对话并流式输出增量内容。
func (a *Agent) ChatStream(ctx context.Context, notebookID, message string, history []ChatMessage, onToken func(string) error) (*ChatResponse, error) {
	chatResult, err := a.buildChatPrompt(ctx, message, history)
	if err != nil {
		return nil, err
	}

	var contentBuilder strings.Builder
	streamingFunc := func(ctx context.Context, chunk []byte) error {
		if len(chunk) == 0 {
			return nil
		}
		text := string(chunk)
		contentBuilder.WriteString(text)
		if onToken != nil {
			return onToken(text)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	response, err := a.provider.GenerateFromSinglePrompt(ctx, a.llm, chatResult.Prompt, llms.WithStreamingFunc(streamingFunc))
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	finalContent := contentBuilder.String()
	if finalContent == "" {
		finalContent = response
	}

	return &ChatResponse{
		Message:   finalContent,
		Sources:   chatResult.Sources,
		SessionID: notebookID,
		Metadata: map[string]interface{}{
			"docs_retrieved": chatResult.DocsRetrieved,
		},
	}, nil
}

type chatPromptResult struct {
	Prompt        string
	Sources       []SourceSummary
	DocsRetrieved int
}

func (a *Agent) buildChatPrompt(ctx context.Context, message string, history []ChatMessage) (*chatPromptResult, error) {
	// Perform similarity search to find relevant sources
	docs, err := a.vectorStore.SimilaritySearch(ctx, message, a.cfg.MaxSources)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	// Build context from retrieved documents
	var contextBuilder strings.Builder
	if len(docs) > 0 {
		contextBuilder.WriteString("来源中的相关信息：\n\n")
		for i, doc := range docs {
			contextBuilder.WriteString(fmt.Sprintf("[来源 %d] %s\n", i+1, doc.PageContent))
			if source, ok := doc.Metadata["source"].(string); ok {
				contextBuilder.WriteString(fmt.Sprintf("来源: %s\n\n", source))
			}
		}
	}

	// Build chat history
	var historyBuilder strings.Builder
	for i, msg := range history {
		if i >= 10 { // Limit history
			break
		}
		role := "用户"
		if msg.Role == "assistant" {
			role = "助手"
		}
		historyBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// Create RAG prompt using f-string format
	promptTemplate := prompts.NewPromptTemplate(
		chatSystemPrompt(),
		[]string{"history", "context", "question"},
	)
	promptTemplate.TemplateFormat = prompts.TemplateFormatFString

	promptValue, err := promptTemplate.Format(map[string]any{
		"history":  historyBuilder.String(),
		"context":  contextBuilder.String(),
		"question": message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// Build source summaries
	sourceSummaries := make([]SourceSummary, 0, len(docs))
	sourceMap := make(map[string]bool)
	for _, doc := range docs {
		if source, ok := doc.Metadata["source"].(string); ok {
			if !sourceMap[source] {
				sourceSummaries = append(sourceSummaries, SourceSummary{
					ID:   source,
					Name: source,
					Type: "file",
				})
				sourceMap[source] = true
			}
		}
	}

	return &chatPromptResult{
		Prompt:        promptValue,
		Sources:       sourceSummaries,
		DocsRetrieved: len(docs),
	}, nil
}

// Slide represents a parsed PPT slide
type Slide struct {
	Style   string
	Content string
}

// ParsePPTSlides parses the LLM output into individual slides
func (a *Agent) ParsePPTSlides(content string) []Slide {
	var slides []Slide

	// 1. Extract style instructions
	style := ""
	styleStart := strings.Index(content, "<STYLE_INSTRUCTIONS>")
	styleEnd := strings.Index(content, "</STYLE_INSTRUCTIONS>")
	if styleStart != -1 && styleEnd > styleStart {
		style = content[styleStart+20 : styleEnd]
	}

	// 2. Split by Slide markers.
	// We look for "Slide X" or "幻灯片 X" with optional Markdown headers
	re := regexp.MustCompile(`(?m)^(?:\s*#{1,6}\s*)?(?:Slide|幻灯片|第\d+张幻灯片|##)\s*\d+[:\s]*.*$`)
	indices := re.FindAllStringIndex(content, -1)

	if len(indices) > 0 {
		for i := 0; i < len(indices); i++ {
			start := indices[i][0]
			end := len(content)
			if i+1 < len(indices) {
				end = indices[i+1][0]
			}

			slideContent := content[start:end]
			// Validation: Must contain at least one of the section markers
			lower := strings.ToLower(slideContent)
			if strings.Contains(lower, "叙事目标") ||
				strings.Contains(lower, "narrative goal") ||
				strings.Contains(lower, "关键内容") {
				slides = append(slides, Slide{
					Style:   style,
					Content: slideContent,
				})
			}
		}
	}

	// 3. If still nothing, try splitting by the required // NARRATIVE GOAL / // 叙事目标
	if len(slides) == 0 {
		// Use a more specific marker for splitting if Slide headers are missing
		marker := "// 叙事目标"
		if !strings.Contains(content, marker) {
			marker = "// NARRATIVE GOAL"
		}

		if strings.Contains(content, marker) {
			parts := strings.Split(content, marker)
			for i := 1; i < len(parts); i++ {
				slides = append(slides, Slide{
					Style:   style,
					Content: marker + parts[i],
				})
			}
		}
	}

	// Final fallback for completely unstructured content
	if len(slides) == 0 {
		slides = append(slides, Slide{Style: style, Content: content})
	}

	return slides
}

// GeneratePodcastScript generates a podcast script from sources
func (a *Agent) GeneratePodcastScript(ctx context.Context, sources []Source, voice string) (string, error) {
	req := &TransformationRequest{
		Type:   "podcast",
		Length: "medium",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateOutline generates an outline from sources
func (a *Agent) GenerateOutline(ctx context.Context, sources []Source) (string, error) {
	req := &TransformationRequest{
		Type:   "outline",
		Length: "detailed",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateFAQ generates an FAQ from sources
func (a *Agent) GenerateFAQ(ctx context.Context, sources []Source) (string, error) {
	req := &TransformationRequest{
		Type:   "faq",
		Length: "comprehensive",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateStudyGuide generates a study guide from sources
func (a *Agent) GenerateStudyGuide(ctx context.Context, sources []Source) (string, error) {
	req := &TransformationRequest{
		Type:   "study_guide",
		Length: "comprehensive",
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// GenerateSummary generates a summary from sources
func (a *Agent) GenerateSummary(ctx context.Context, sources []Source, length string) (string, error) {
	req := &TransformationRequest{
		Type:   "summary",
		Length: length,
		Format: "markdown",
	}

	resp, err := a.GenerateTransformation(ctx, req, sources)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// callDeepInsight executes the DeepInsight CLI tool and returns the generated report
func (a *Agent) callDeepInsight(ctx context.Context, summary string) (string, error) {
	// Create a temporary file for the report output
	tmpFile := "./tmp/deepinsight_report_" + fmt.Sprintf("%d", time.Now().Unix()) + ".md"

	// Execute DeepInsight command
	// DeepInsight -o report.md "summary text"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	output, err := execCommandContext(ctx, "./DeepInsight", "-o", tmpFile, escapeShellArg(summary))
	if err != nil {
		golog.Infof("failed to exec DeepInsight: err=%v, output=%s", err, output)
		return "", fmt.Errorf("DeepInsight command failed: %w, output: %s", err, output)
	}

	// Read the generated report
	reportContent, err := execCommandContext(ctx, "/bin/cat", tmpFile)
	if err != nil {
		golog.Infof("failed to read DeepInsight report: err=%v, output=%s", err, output)
		return "", fmt.Errorf("failed to read DeepInsight report: %w", err)
	}

	// Clean up temp file
	_, _ = execCommandContext(context.Background(), "/bin/rm", "-f", tmpFile)

	return reportContent, nil
}

// escapeShellArg escapes a shell argument to prevent injection
func escapeShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
}

// execCommandContext is a helper to execute commands with context
func execCommandContext(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}
