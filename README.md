# Notex

English | [中文](./README_CN.md) | [繁體中文](./README_zh-tw.md)

<div align="center">

**A privacy-first, open-source alternative to NotebookLM**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](./LICENSE)

An AI-powered knowledge management application that lets you create intelligent notebooks from your documents.

**Project URL:** https://github.com/smallnest/notex

![](docs/note2.png)

</div>

- Python clone: [pynotex](https://github.com/Beeta/pynotex)

## ✨ Features

- 📚 **Multiple Source Types** - Upload PDFs, text files, Markdown, DOCX, and HTML documents
- 🤖 **AI-Powered Chat** - Ask questions and get answers based on your sources
- ✨ **Multiple Transformations** - Generate summaries, FAQs, study guides, outlines, timelines, glossaries, quizzes, mindmaps, infographics and podcast scripts
- 📊 **Infographic Generation** - Create beautiful, hand-drawn style infographics from your content using Google's Gemini Nano Banana
- 🎙️ **Podcast Generation** - Create engaging podcast scripts from your content
- 💾 **Full Privacy** - Local SQLite storage, optional cloud backends
- 🔄 **Multi-Model Support** - Works with OpenAI, Ollama, and other compatible APIs
- 🎨 **Academic Brutalist Design** - Distinctive, research-focused interface

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or later
- An LLM API key (OpenAI) or Ollama running locally

### Installation

```bash
# Clone the repository
git clone https://github.com/smallnest/notex.git
cd notex

# Install dependencies
go mod tidy

# Run the server
go run . -server
```

Open your browser to `http://localhost:8080`

## ⚙️ Configuration

Notex uses environment variables for configuration. The recommended way to configure the application is to create a `.env` file.

### Step 1: Create Configuration File

Copy the example configuration file to create your local configuration:

```bash
cp .env.example .env
```

### Step 2: Configure Your LLM Provider

Edit the `.env` file and configure **one** of the following LLM providers:

#### Option A: Using OpenAI (Cloud-based)

OpenAI provides high-quality models but requires an API key and charges per usage.

1. Get an API key from [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys)
2. Edit `.env` and configure:

```env
# OpenAI Configuration
OPENAI_API_KEY=sk-your-actual-api-key-here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
EMBEDDING_MODEL=text-embedding-3-small
```

**Available OpenAI Models:**

- `gpt-4o-mini` - Fast and cost-effective (recommended)
- `gpt-4o` - Most capable
- `gpt-3.5-turbo` - Legacy option

**Tips:**

- You can also use compatible OpenAI APIs like Azure OpenAI or other providers by changing `OPENAI_BASE_URL`
- For example, to use DeepSeek: `OPENAI_BASE_URL=https://api.deepseek.com/v1` and `OPENAI_MODEL=deepseek-chat`

#### Option B: Using Ollama (Local, Free)

Ollama runs locally on your machine and is completely free, but requires a capable computer.

1. Install Ollama from [https://ollama.com](https://ollama.com)
2. Pull a model (e.g., `ollama pull llama3.2`)
3. Start Ollama: `ollama serve`
4. Edit `.env` and configure:

```env
# Ollama Configuration
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2
```

**Available Ollama Models:**

- `llama3.2` - Recommended balance of speed and quality
- `qwen2.5` - Excellent for Chinese content
- `mistral` - Good English performance
- `codellama` - Specialized for code

**Tips:**

- Ollama models run entirely on your machine - no data leaves your computer
- Make sure Ollama is running before starting Notex
- Larger models require more RAM and CPU

### Step 3: Optional Google Gemini (for Infographics)

To use the infographic generation feature with Google's Gemini Nano Banana:

```env
GOOGLE_API_KEY=your-google-api-key-here
```

Get your key from [https://makersuite.google.com/app/apikey](https://makersuite.google.com/app/apikey)

### Step 4: Run the Application

After configuring your `.env` file, simply run:

```bash
go run . -server
```

The application will automatically load your `.env` configuration and start at `http://localhost:8080`

### Build and Run (Optional)

If you prefer to build a binary instead of using `go run`:

```bash
go build -o notex .
./notex -server
```

## 📖 Usage

### Creating Notebooks

1. Click "New Notebook" in the header
2. Enter a name and optional description
3. Click "Create Notebook"

### Adding Sources

You can add content to your notebook in three ways:

**File Upload**

- Click the "+" button in the Sources panel
- Drag and drop or browse for files
- Supported: PDF, TXT, MD, DOCX, HTML

**Paste Text**

- Select the "Text" tab
- Enter a title and paste your content

**From URL**

- Select the "URL" tab
- Enter the URL and optional title

### Chatting with Sources

1. Switch to the "CHAT" tab
2. Ask questions about your content
3. Responses include references to relevant sources

### Transformations

Click any transformation card to generate:

| Transformation | Description                                                  |
| -------------- | ------------------------------------------------------------ |
| 📝 Summary     | Condensed overview of your sources                           |
| ❓ FAQ         | Common questions and answers                                 |
| 📚 Study Guide | Educational material with learning objectives                |
| 🗂️ Outline     | Hierarchical structure of topics                             |
| 🎙️ Podcast     | Conversational script for audio content                      |
| 📅 Timeline    | Chronological events from sources                            |
| 📖 Glossary    | Key terms and definitions                                    |
| ✍️ Quiz        | Assessment questions with answer key                         |
| 📊 Infographic | Hand-drawn style visual representation of your content       |
| 🧠 Mindmap     | Visual hierarchical diagram of your sources using Mermaid.js |

Or use the custom prompt field for any other transformation.

### Additional Configuration Options

For advanced users, the `.env` file supports additional configuration options:

```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Vector Store (default: sqlite)
# Options: sqlite, memory, supabase, postgres, redis
VECTOR_STORE_TYPE=sqlite

# RAG Processing
MAX_SOURCES=5          # Maximum sources to retrieve for context
CHUNK_SIZE=1000        # Document chunk size for processing
CHUNK_OVERLAP=200      # Overlap between chunks

# Document Conversion
ENABLE_MARKITDOWN=true  # Use Microsoft markitdown for better PDF/DOCX conversion

# Podcast Generation
ENABLE_PODCAST=true
PODCAST_VOICE=alloy    # Options: alloy, echo, fable, onyx, nova, shimmer

# Feature Flags
ALLOW_DELETE=true
ALLOW_MULTIPLE_NOTES_OF_SAME_TYPE=true

# S3 / Ceph Storage
S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1   # default us-east-1
S3_ACCESS_KEY=yourkey
S3_SECRET_KEY=yoursecret
S3_BUCKET=notex-uploads
S3_FORCE_PATH_STYLE=true  # usually true for Ceph/MinIO
S3_SKIP_TLS_VERIFY=true # default false

```

## 🔧 Development

### Running Tests

```bash
go test -v ./...
```

### Building

```bash
go build -o notex .
```

### Code Quality

```bash
# Format
go fmt ./...

# Lint
golangci-lint run

# Vet
go vet ./...
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

Apache License 2.0 - see [LICENSE](./LICENSE) for details.

## 🙏 Acknowledgments

- Inspired by [Google's NotebookLM](https://notebooklm.google.com/)
- Built with [LangGraphGo](https://github.com/smallnest/langgraphgo)
- Inspired by [notex](https://github.com/lfnovo/notex)

## 📞 Support

- Report issues on [GitHub](https://github.com/smallnest/notex/issues)
- Join discussions in the [Notex community](https://github.com/smallnest/notex/discussions)

---

**Notex** - A privacy-first, open-source alternative to NotebookLM
https://github.com/smallnest/notex
