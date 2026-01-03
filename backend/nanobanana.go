package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kataras/golog"
	"google.golang.org/genai"
)

// GenerateInfographImage generates an infographic image using the Nano Banana Pro SDK
func (a *Agent) GenerateInfographImage(ctx context.Context, prompt string) (string, error) {
	if a.cfg.GoogleAPIKey == "" {
		golog.Errorf("google_api_key is not set")
		return "", fmt.Errorf("google_api_key is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  a.cfg.GoogleAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %w", err)
	}

	prompt += "\n\n **注意：无论来源是什么语言，请务必使用中文**\n"

	// Using gemini-3-pro-image-preview as per the new documentation example
	model := "gemini-3-pro-image-preview"
	golog.Infof("generating infographic with model %s using GenerateContent...", model)

	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	resp, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		golog.Errorf("failed to generate content: %v", err)
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		golog.Errorf("no candidates returned by the model")
		return "", fmt.Errorf("no candidates generated")
	}

	var imageData []byte
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			imageData = part.InlineData.Data
			break
		}
	}

	if len(imageData) == 0 {
		golog.Errorf("no image data found in the response parts")
		return "", fmt.Errorf("no image data in response")
	}

	golog.Infof("image data received successfully, saving...")

	// Save the image
	fileName := fmt.Sprintf("infograph_%d.png", time.Now().UnixNano())
	uploadDir := "./data/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	filePath := filepath.Join(uploadDir, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		golog.Errorf("failed to save image to %s: %v", filePath, err)
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	golog.Infof("infographic saved to %s", filePath)
	return filePath, nil
}
