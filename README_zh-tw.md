# Notex

[English](./README.md) | [中文](./README_CN.md) | 繁體中文

<div align="center">

**隱私優先的開源 NotebookLM 替代方案**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](./LICENSE)

一個由 AI 驅動的知識管理應用程式，讓你從文件中創建智慧筆記本。

**專案網址：** https://github.com/smallnest/notex

![](docs/note2.png)

</div>

- Python 版本：[pynotex](https://github.com/Beeta/pynotex)

## ✨ 功能特色

- 📚 **多種來源類型** - 上傳 PDF、文字檔、Markdown、DOCX 和 HTML 文件
- 🤖 **AI 驅動對話** - 基於你的來源資料提問並獲得解答
- ✨ **多種轉換功能** - 生成摘要、常見問題、學習指南、大綱、時間軸、詞彙表、測驗、心智圖、資訊圖表和播客腳本
- 📊 **資訊圖表生成** - 使用 Google 的 Gemini Nano Banana 從你的內容創建美麗的手繪風格資訊圖表
- 🎙️ **播客生成** - 從你的內容創建引人入勝的播客腳本
- 💾 **完全隱私** - 本地 SQLite 儲存，可選雲端後端
- 🔄 **多模型支援** - 與 OpenAI、Ollama 和其他相容的 API 協同工作
- 🎨 **學術粗野主義設計** - 獨特的研究導向介面

## 🚀 快速開始

### 前置需求

- Go 1.23 或更新版本
- LLM API 金鑰（OpenAI）或在本地執行的 Ollama

### 安裝

```bash
# 複製儲存庫
git clone https://github.com/smallnest/notex.git
cd notex

# 安裝依賴套件
go mod tidy

# 執行伺服器
go run . -server
```

在瀏覽器中開啟 `http://localhost:8080`

## ⚙️ 配置

Notex 使用環境變數進行配置。推薦的配置方式是創建 `.env` 檔案。

### 步驟 1：創建配置檔案

複製範例配置檔案以創建你的本地配置：

```bash
cp .env.example .env
```

### 步驟 2：配置你的 LLM 提供商

編輯 `.env` 檔案並配置以下 LLM 提供商**其中一個**：

#### 選項 A：使用 OpenAI（雲端）

OpenAI 提供高品質模型，但需要 API 金鑰並按使用量收費。

1. 從 [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys) 取得 API 金鑰
2. 編輯 `.env` 並配置：

```env
# OpenAI 配置
OPENAI_API_KEY=sk-your-actual-api-key-here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
EMBEDDING_MODEL=text-embedding-3-small
```

**可用的 OpenAI 模型：**

- `gpt-4o-mini` - 快速且經濟實惠（推薦）
- `gpt-4o` - 最強大的模型
- `gpt-3.5-turbo` - 舊版選項

**提示：**

- 你也可以透過更改 `OPENAI_BASE_URL` 來使用相容的 OpenAI API，如 Azure OpenAI 或其他提供商
- 例如，要使用 DeepSeek：`OPENAI_BASE_URL=https://api.deepseek.com/v1` 和 `OPENAI_MODEL=deepseek-chat`

#### 選項 B：使用 Ollama（本地、免費）

Ollama 在你的機器上本地執行，完全免費，但需要性能足夠的電腦。

1. 從 [https://ollama.com](https://ollama.com) 安裝 Ollama
2. 拉取模型（例如：`ollama pull llama3.2`）
3. 啟動 Ollama：`ollama serve`
4. 編輯 `.env` 並配置：

```env
# Ollama 配置
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2
```

**可用的 Ollama 模型：**

- `llama3.2` - 推薦的速度與品質平衡
- `qwen2.5` - 優秀的中文內容處理
- `mistral` - 良好的英文效能
- `codellama` - 專門用於程式碼

**提示：**

- Ollama 模型完全在你的機器上執行 - 資料不會離開你的電腦
- 在啟動 Notex 之前確保 Ollama 正在執行
- 較大的模型需要更多的記憶體和 CPU

### 步驟 3：選用 Google Gemini（用於資訊圖表）

要使用 Google 的 Gemini Nano Banana 資訊圖表生成功能：

```env
GOOGLE_API_KEY=your-google-api-key-here
```

從 [https://makersuite.google.com/app/apikey](https://makersuite.google.com/app/apikey) 取得你的金鑰

### 步驟 4：執行應用程式

配置好 `.env` 檔案後，只需執行：

```bash
go run . -server
```

應用程式將自動載入你的 `.env` 配置並在 `http://localhost:8080` 啟動

### 建置並執行（選用）

如果你偏好建置二進位檔案而不是使用 `go run`：

```bash
go build -o notex .
./notex -server
```

## 📖 使用方式

### 創建筆記本

1. 點擊標題中的「New Notebook」
2. 輸入名稱和選用的描述
3. 點擊「Create Notebook」

### 新增來源

你可以透過三種方式將內容新增到筆記本：

**檔案上傳**

- 點擊來源面板中的「+」按鈕
- 拖放或瀏覽檔案
- 支援格式：PDF、TXT、MD、DOCX、HTML

**貼上文字**

- 選擇「Text」分頁
- 輸入標題並貼上你的內容

**從 URL**

- 選擇「URL」分頁
- 輸入 URL 和選用的標題

### 與來源對話

1. 切換到「CHAT」分頁
2. 詢問關於你的內容的問題
3. 回應包含相關來源的引用

### 轉換功能

點擊任何轉換卡片以生成：

| 轉換功能    | 描述                               |
| ----------- | ---------------------------------- |
| 📝 摘要     | 你的來源的濃縮概覽                 |
| ❓ 常見問題 | 常見問題和解答                     |
| 📚 學習指南 | 包含學習目標的教育材料             |
| 🗂️ 大綱     | 主題的階層結構                     |
| 🎙️ 播客     | 音訊內容的對話腳本                 |
| 📅 時間軸   | 來源中的時間順序事件               |
| 📖 詞彙表   | 關鍵術語和定義                     |
| ✍️ 測驗     | 包含答案的評估問題                 |
| 📊 資訊圖表 | 你的內容的手繪風格視覺化呈現       |
| 🧠 心智圖   | 使用 Mermaid.js 的來源視覺化階層圖 |

或使用自訂提示欄位進行任何其他轉換。

### 額外配置選項

對於進階使用者，`.env` 檔案支援額外的配置選項：

```env
# 伺服器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 向量儲存（預設：sqlite）
# 選項：sqlite、memory、supabase、postgres、redis
VECTOR_STORE_TYPE=sqlite

# RAG 處理
MAX_SOURCES=5          # 檢索上下文的最大來源數
CHUNK_SIZE=1000        # 文件區塊大小以供處理
CHUNK_OVERLAP=200      # 區塊之間的重疊

# 文件轉換
ENABLE_MARKITDOWN=true  # 使用 Microsoft markitdown 以獲得更好的 PDF/DOCX 轉換

# 播客生成
ENABLE_PODCAST=true
PODCAST_VOICE=alloy    # 選項：alloy、echo、fable、onyx、nova、shimmer

# 功能標誌
ALLOW_DELETE=true
ALLOW_MULTIPLE_NOTES_OF_SAME_TYPE=true

# S3 / Ceph 對象儲存，檔案上傳
S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1   # default `us-east-1`
S3_ACCESS_KEY=yourkey
S3_SECRET_KEY=yoursecret
S3_BUCKET=notex-uploads
S3_FORCE_PATH_STYLE=true  # usually true for Ceph/MinIO
S3_SKIP_TLS_VERIFY=true # default false
```

## 🔧 開發

### 執行測試

```bash
go test -v ./...
```

### 建置

```bash
go build -o notex .
```

### 程式碼品質

```bash
# 格式化
go fmt ./...

# 語法檢查
golangci-lint run

# 審查
go vet ./...
```

## 🤝 貢獻

歡迎貢獻！請隨時提交 Pull Request。

## 📄 授權

Apache License 2.0 - 詳見 [LICENSE](./LICENSE)。

## 🙏 致謝

- 靈感來自 [Google's NotebookLM](https://notebooklm.google.com/)
- 使用 [LangGraphGo](https://github.com/smallnest/langgraphgo) 建置
- 靈感來自 [notex](https://github.com/lfnovo/notex)

## 📞 支援

- 在 [GitHub](https://github.com/smallnest/notex/issues) 上回報問題
- 加入 [Notex 社群](https://github.com/smallnest/notex/discussions)的討論

---

**Notex** - 隱私優先的開源 NotebookLM 替代方案
https://github.com/smallnest/notex
