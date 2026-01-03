package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	ollamallm "github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

// Agent handles AI operations for generating notes and chat responses
type Agent struct {
	vectorStore *VectorStore
	llm         llms.Model
	cfg         Config
}

// NewAgent creates a new agent
func NewAgent(cfg Config, vectorStore *VectorStore) (*Agent, error) {
	llm, err := createLLM(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	return &Agent{
		vectorStore: vectorStore,
		llm:         llm,
		cfg:         cfg,
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
	promptTemplate := a.getTransformationPrompt(req)

	prompt := prompts.NewPromptTemplate(
		promptTemplate,
		[]string{"sources", "type", "length", "format", "prompt"},
	)
	prompt.TemplateFormat = prompts.TemplateFormatFString

	// Generate response
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	response, err := llms.GenerateFromSinglePrompt(ctx, a.llm, promptValue)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
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

// getTransformationPrompt returns the prompt template for each transformation type
func (a *Agent) getTransformationPrompt(req *TransformationRequest) string {
	switch req.Type {
	case "summary":
		return `你是一个擅长创建综合摘要的专家。请根据以下来源，以{format}格式创建一个{length}摘要。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

请提供一个结构良好的摘要，捕捉来源中的关键信息、主要主题和重要细节。`

	case "faq":
		return `你是一个擅长创建常见问题解答（FAQ）文档的专家。请根据以下来源，以{format}格式生成一个全面的FAQ。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

创建10-15个常见问题及其详细解答，涵盖来源中的主要主题和信息。`

	case "study_guide":
		return `你是一个教育专家。请根据以下来源，以{format}格式创建一个全面的学习指南。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

学习指南应包括：
1. 学习目标
2. 关键概念和定义
3. 重要主题和议题
4. 学习问题和练习
5. 要点总结

请针对{length}的学习课程进行格式化。`

	case "outline":
		return `你是一个擅长创建结构化大纲的专家。请根据以下来源，以{format}格式创建一个详细的层级大纲。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

大纲应：
- 使用适当的层级结构（I, A, 1, a）
- 涵盖所有主要主题和子主题
- 包含主要部分的简要说明
- 详细程度为{length}`

	case "podcast":
		return `你是一个播客脚本编剧。请根据以下来源创建一个引人入胜的播客脚本。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

脚本应：
- 具有对话性和吸引力
- 涵盖来源中的主要主题
- 包括两位主持人讨论材料
- 口语时长约为10-15分钟
- 包含自然的过渡和提问
- 有清晰的开场白和结束语

请将其格式化为带有演讲者标签（主持人1，主持人2）和[括号]中舞台指示的播客脚本。`

	case "timeline":
		return `你是一个擅长创建按时间顺序排列的时间线的专家。请根据以下来源，以{format}格式创建一个时间线。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

按时间顺序提取和组织事件，包括：
- 日期或时间段
- 事件描述
- 涉及的关键人物
- 每个事件的重要性`

	case "glossary":
		return `你是一个擅长创建术语表的专家。请根据以下来源，以{format}格式创建一个全面的术语表。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

包括：
- 重要术语和概念
- 清晰简洁的定义
- 来源中的上下文
- 相关术语之间的交叉引用`

	case "quiz":
		return `你是一个创建评估材料的教育家。请根据以下来源，以{format}格式创建一个测验。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

测验应包括：
- 混合题型（多项选择、判断正误、简答）
- 不同难度的问题
- 答案
- 测试理解力而非仅仅是记忆力的问题

创建一个包含10-20个问题的{length}测验。`

	case "mindmap":
		return `你是一位资深的信息架构师和知识管理专家。请将【文本内容】提炼并转换为 Mermaid.js 的 mindmap 格式。
**注意：无论来源是什么语言，请务必使用中文进行回复。**

# 样式规范：
1. **中心主题**：必须使用 root((内容)) 格式（圆圈）。
2. **主要分支**：使用 (内容) 格式（圆角矩形）。
3. **细节节点**：使用 [内容] 格式（普通矩形）或直接写文字。

# 严格逻辑规范：
1. **仅限 mindmap 语法**：严禁使用 graph, LR, --> 等字符。
2. **内容安全**：节点内容必须精炼（10字以内），严禁包含引号。
3. **严禁解释**：只输出以 ` + "```mermaid" + ` 开头和以 ` + "```" + ` 结尾的代码块。

来源：
{sources}

# 示例：
` + "```mermaid" + `
mindmap
  root((核心主题))
    (主要分支A)
      [细节1]
      [细节2]
    (主要分支B)
      [细节3]
` + "```" + `
`

	case "infograph":
		return `# Role
你是一位世界顶级的数据可视化设计师和信息图专家。你的任务是将复杂的文本信息转化为直观、吸引人且准确的视觉设计方案。

# Task
阅读所附文本，设计一张信息图（Infographic）。不要进行总结，而是描述这张图应该长什么样。你的输出将被直接用作 DALL-E 3 的绘画提示词。
**注意：无论来源是什么语言，请务必使用中文进行回复。确保信息图中的所有文本内容（如标题、标签、数据点）都使用中文。不要使用 ` + "```markdown" + ` 标记包裹输出。**

# Design Guidelines
1.  **核心信息提炼**：找出文本中最重要的 3-5 个数据点、流程步骤或对比项。
2.  **视觉隐喻**：使用形象的比喻。例如，讲网络安全用“盾牌和锁”，讲增长用“火箭或上升的箭头”。
3.  **布局结构**：明确定义图的结构（例如：“从左到右的流程图”、“分成两半的对比图”、“中心辐射图”）。
4.  **文本限制**：信息图中的文字必须极简。只保留标题、关键数据和极短的标签。
5.  **风格**：插画或手绘感，使用柔和的插画或轻松的手绘笔触，以增强亲和力和友好度。

# Output Format (DALL-E 3 Prompt style)
Start with "Infographic illustration created in a soft, hand-drawn digital art style with friendly and approachable vibes."
[描述整体布局和背景风格]
[详细描述主要视觉元素 1，包含其图标、颜色、位置和附带的文字标签]
[详细描述主要视觉元素 2，...]
[详细描述连接元素（如箭头、线条）]
[描述整体标题和配色方案]
End with "The background is a clean, light gradient suitable for a professional presentation."

# Input Text
{sources}
`

	case "custom":
		return `你是一个有用的助手。根据以下来源和自定义请求，生成请求的内容。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

自定义请求：
{prompt}

请以{format}格式生成内容，保持{length}。`

	default:
		return `你是一个有用的助手。根据以下来源，以{format}格式提供一个{type}。
**注意：无论来源是什么语言，请务必使用中文进行回复。不要使用 ` + "```markdown" + ` 标记包裹输出。**

来源：
{sources}

生成{length}内容。`
	}
}

// Chat performs a chat query with RAG
func (a *Agent) Chat(ctx context.Context, notebookID, message string, history []ChatMessage) (*ChatResponse, error) {
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
	systemPrompt := `你是一个笔记本应用程序的有用人工智能助手。根据提供的上下文和聊天历史记录回答用户的问题。
**无论来源文件是什么语言，请务必使用中文回答用户的问题。不要使用 ` + "```markdown" + ` 标记包裹输出。**
如果上下文中没有足够的信息，请说明情况并提供一般性的回答。

聊天历史记录：
{history}

上下文：
{context}

用户问题：{question}

请提供有用的、准确的回答。当引用来源中的信息时，请提及信息来自哪个来源。`

	promptTemplate := prompts.NewPromptTemplate(
		systemPrompt,
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

	// Generate response
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	response, err := llms.GenerateFromSinglePrompt(ctx, a.llm, promptValue)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
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

	return &ChatResponse{
		Message:   response,
		Sources:   sourceSummaries,
		SessionID: notebookID,
		Metadata: map[string]interface{}{
			"docs_retrieved": len(docs),
		},
	}, nil
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
