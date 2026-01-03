package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kataras/golog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
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
		fmt.Printf("Notex v%s\n", Version)
		fmt.Println("A privacy-first, open-source alternative to NotebookLM")
		fmt.Println("Powered by LangGraphGo")
		os.Exit(0)
	}

	defer func() {
		if err := recover(); err != nil {
			golog.Error("recover:", err)
			buf := make([]byte, 8192)
			n := runtime.Stack(buf, true)
			golog.Error("stack:", string(buf[:n]))
		}
	}()

	golog.SetTimeFormat("2006/01/02 15:04:05.000")
	logFiles := "./logs/notex.log.%Y%m%d"
	w, err := rotatelogs.New(
		logFiles,
		rotatelogs.WithLinkName("./logs/notex.log"),
		rotatelogs.WithMaxAge(time.Duration(7)*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour))
	if err != nil {
		golog.Fatal(err)
	}
	defer w.Close()
	golog.SetOutput(w)

	// Load and validate configuration
	cfg := backend.LoadConfig()
	if err := backend.ValidateConfig(cfg); err != nil {
		golog.Fatalf("configuration error: %v\n\n"+
			"Required environment variables:\n"+
			"  - OPENAI_API_KEY (for OpenAI) or\n"+
			"  - OLLAMA_BASE_URL (for local Ollama)\n\n"+
			"Optional:\n"+
			"  - VECTOR_STORE_TYPE (default: sqlite)\n"+
			"  - STORE_PATH (default: ./data/checkpoints.db)\n"+
			"  - SERVER_PORT (default: 8080)\n"+
			"Error: %v", err, err)
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
		golog.Fatalf("failed to create server: %v", err)
	}

	golog.Infof("version:     %s", Version)
	golog.Infof("server:      http://%s:%s", cfg.ServerHost, cfg.ServerPort)
	golog.Infof("llm:         %s", cfg.OpenAIModel)
	golog.Infof("vector store: %s", cfg.VectorStoreType)

	if err := server.Start(); err != nil {
		golog.Fatalf("server error: %v", err)
	}
}

func runIngestMode(ctx context.Context, cfg backend.Config, filePath, notebookName string) {
	golog.Infof("📂 ingesting file: %s...", filePath)

	// Initialize vector store
	vectorStore, err := backend.NewVectorStore(cfg)
	if err != nil {
		golog.Fatalf("failed to initialize vector store: %v", err)
	}

	// Initialize store
	store, err := backend.NewStore(cfg)
	if err != nil {
		golog.Fatalf("failed to initialize store: %v", err)
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
			golog.Fatalf("failed to create notebook: %v", err)
		}
		notebookID = nb.ID
		golog.Infof("📓 created notebook: %s", notebookName)
	}

	// Extract content
	content, err := vectorStore.ExtractDocument(ctx, filePath)
	if err != nil {
		golog.Fatalf("extraction failed: %v", err)
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
		golog.Fatalf("failed to create source: %v", err)
	}

	// Ingest document
	if err := vectorStore.IngestText(ctx, source.Name, content); err != nil {
		golog.Fatalf("ingestion failed: %v", err)
	}

	golog.Infof("✅ ingestion complete!")
	golog.Infof("📓 notebook: %s (ID: %s)", notebookName, notebookID)
}

func printUsage() {
	fmt.Println("Notex - Privacy-first AI notebook")
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
