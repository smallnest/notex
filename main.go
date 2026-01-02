package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/smallnest/notex/backend"
)

var Version = "1.0.0"

func main() {
	// Command line flags
	serverMode := flag.Bool("server", false, "Run in HTTP server mode")
	ingestFile := flag.String("ingest", "", "Path to a file to ingest")
	notebookName := flag.String("notebook", "", "Notebook name (for ingest)")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		fmt.Printf("Open Notebook v%s\n", Version)
		fmt.Println("A privacy-first, open-source alternative to NotebookLM")
		fmt.Println("Powered by LangGraphGo")
		os.Exit(0)
	}

	// Load and validate configuration
	cfg := backend.LoadConfig()
	if err := backend.ValidateConfig(cfg); err != nil {
		log.Fatalf("Configuration error: %v\n\n"+
			"Required environment variables:\n"+
			"  - OPENAI_API_KEY (for OpenAI) or\n"+
			"  - OLLAMA_BASE_URL (for local Ollama)\n\n"+
			"Optional:\n"+
			"  - VECTOR_STORE_TYPE (default: sqlite)\n"+
			"  - STORE_PATH (default: ./data/checkpoints.db)\n"+
			"  - SERVER_PORT (default: 8080)\n"+
			"Error: %v", err)
	}

	ctx := context.Background()

	switch {
	case *serverMode:
		// Server mode
		runServerMode(cfg)

	case *ingestFile != "":
		// Ingest mode
		if *notebookName == "" {
			*notebookName = "Default Notebook"
		}
		runIngestMode(ctx, cfg, *ingestFile, *notebookName)

	default:
		printUsage()
	}
}

func runServerMode(cfg backend.Config) {
	server, err := backend.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("\n")
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║           📓 Open Notebook - LangGraphGo Edition          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Printf("\n")
	fmt.Printf("Version:     %s\n", Version)
	fmt.Printf("Server:      http://%s:%s\n", cfg.ServerHost, cfg.ServerPort)
	fmt.Printf("LLM:         %s\n", cfg.OpenAIModel)
	fmt.Printf("Vector Store: %s\n", cfg.VectorStoreType)
	fmt.Printf("\n")
	fmt.Println("Features:")
	fmt.Println("  📚 Multiple source types (PDF, TXT, MD, DOCX, HTML)")
	fmt.Println("  🤖 AI-powered chat with your documents")
	fmt.Println("  ✨ Multiple transformations (summary, FAQ, study guide, outline)")
	fmt.Println("  🎙️  Podcast generation")
	fmt.Println("  💾 Full privacy - local storage")
	fmt.Println("  🔄 Multi-model support (OpenAI, Ollama)")
	fmt.Println("\nPress Ctrl+C to stop")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runIngestMode(ctx context.Context, cfg backend.Config, filePath, notebookName string) {
	fmt.Printf("📂 Ingesting file: %s...\n", filePath)

	// Initialize vector store
	vectorStore, err := backend.NewVectorStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize vector store: %v", err)
	}

	// Initialize store
	store, err := backend.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Create or get notebook
	notebooks, _ := store.ListNotebooks(ctx)
	var notebookID string
	for _, nb := range notebooks {
		if nb.Name == notebookName {
			notebookID = nb.ID
			break
		}
	}

	if notebookID == "" {
		nb, err := store.CreateNotebook(ctx, notebookName, "Created by ingest mode", nil)
		if err != nil {
			log.Fatalf("Failed to create notebook: %v", err)
		}
		notebookID = nb.ID
		fmt.Printf("📓 Created notebook: %s\n", notebookName)
	}

	// Extract content
	content, err := vectorStore.ExtractDocument(ctx, filePath)
	if err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}

	// Create source in database
	fileInfo, _ := os.Stat(filePath)
	source := &backend.Source{
		NotebookID: notebookID,
		Name:       filepath.Base(filePath),
		Type:       "file",
		FileName:   filepath.Base(filePath),
		FileSize:   fileInfo.Size(),
		Content:    content,
		Metadata:   map[string]interface{}{"path": filePath},
	}

	if err := store.CreateSource(ctx, source); err != nil {
		log.Fatalf("Failed to create source: %v", err)
	}

	// Ingest document
	if err := vectorStore.IngestText(ctx, source.Name, content); err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}

	fmt.Println("✅ Ingestion complete!")
	fmt.Printf("📓 Notebook: %s (ID: %s)\n", notebookName, notebookID)
}

func printUsage() {
	fmt.Println("Open Notebook - Privacy-first AI notebook")
	fmt.Println("\nUsage:")
	fmt.Println("  open-notebook [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -server          Start the web server")
	fmt.Println("  -ingest <file>   Ingest a file into the vector store")
	fmt.Println("  -notebook <name> Notebook name for ingest (default: 'Default Notebook')")
	fmt.Println("  -version         Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  # Start web server")
	fmt.Println("  open-notebook -server")
	fmt.Println("\n  # Ingest a file")
	fmt.Println("  open-notebook -ingest document.pdf -notebook 'My Notes'")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  OPENAI_API_KEY      Your OpenAI API key")
	fmt.Println("  OLLAMA_BASE_URL     Ollama server URL (default: http://localhost:11434)")
	fmt.Println("  OPENAI_MODEL        Model name (default: gpt-4o-mini)")
	fmt.Println("  VECTOR_STORE_TYPE   Vector store type (default: sqlite)")
	fmt.Println("  SERVER_PORT         Server port (default: 8080)")
	fmt.Println("\nFor more information, visit: https://github.com/smallnest/langgraphgo")
}
