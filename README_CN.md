# Notex

中文 | [English](./README.md)

<div align="center">

**注重隐私的开源 NotebookLM 替代方案**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](./LICENSE)

AI 驱动知识管理应用程序，让您从文档中创建智能笔记本。

**项目地址：** https://github.com/smallnest/notex

![](docs/note2.png)

</div>

- Python 版: [pynotex](https://github.com/Beeta/pynotex)

## ✨ 特性

- 📚 **多种来源类型** - 支持上传 PDF、文本文件、Markdown、DOCX 和 HTML 文档
- 🤖 **AI 驱动对话** - 基于您的来源提问并获得答案
- ✨ **多种转换** - 生成摘要、FAQ、学习指南、大纲、时间线、词汇表、测验、思维导图、信息图和播客脚本
- 📊 **信息图生成** - 使用 Google Gemini Nano Banana 从您的内容创建精美的手绘风格信息图
- 🎙️ **播客生成** - 从您的内容创建引人入胜的播客脚本
- 💾 **完全隐私** - 本地 SQLite 存储，可选云端后端
- 🔄 **多模型支持** - 兼容 OpenAI、Ollama 和其他兼容 API
- 🎨 **学术野兽派设计** - 独特的研究专注型界面

## 🚀 快速开始

### 前置要求

- Go 1.23 或更高版本
- LLM API 密钥 (OpenAI) 或本地运行的 Ollama

### 安装

```bash
# 克隆仓库
git clone https://github.com/smallnest/notex.git
cd notex

# 安装依赖
go mod tidy

# 运行服务器
go run . -server
```

在浏览器中打开 `http://localhost:8080`

## ⚙️ 配置

Notex 使用环境变量进行配置。推荐的方式是创建 `.env` 文件来配置应用。

### 步骤 1：创建配置文件

复制示例配置文件来创建本地配置：

```bash
cp .env.example .env
```

### 步骤 2：配置您的 LLM 提供商

编辑 `.env` 文件，配置以下 LLM 提供商中的**一个**：

#### 选项 A：使用 OpenAI（云端）

OpenAI 提供高质量的模型，但需要 API 密钥并按使用量收费。

1. 从 [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys) 获取 API 密钥
2. 编辑 `.env` 并配置：

```env
# OpenAI 配置
OPENAI_API_KEY=sk-your-actual-api-key-here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
EMBEDDING_MODEL=text-embedding-3-small
```

**可用的 OpenAI 模型：**

- `gpt-4o-mini` - 快速且经济实惠（推荐）
- `gpt-4o` - 最强大
- `gpt-3.5-turbo` - 旧版本选项

**小贴士：**

- 您也可以通过修改 `OPENAI_BASE_URL` 来使用 Azure OpenAI 或其他兼容 OpenAI 的 API
- 例如，使用 DeepSeek：`OPENAI_BASE_URL=https://api.deepseek.com/v1` 和 `OPENAI_MODEL=deepseek-chat`

#### 选项 B：使用 Ollama（本地、免费）

Ollama 在您的本地机器上运行，完全免费，但需要一台性能较好的电脑。

1. 从 [https://ollama.com](https://ollama.com) 安装 Ollama
2. 拉取一个模型（例如：`ollama pull llama3.2`）
3. 启动 Ollama：`ollama serve`
4. 编辑 `.env` 并配置：

```env
# Ollama 配置
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2
```

**可用的 Ollama 模型：**

- `llama3.2` - 速度和质量的推荐平衡
- `qwen2.5` - 中文内容效果优秀
- `mistral` - 英文性能良好
- `codellama` - 代码专用

**小贴士：**

- Ollama 模型完全在您的机器上运行 - 数据不会离开您的电脑
- 确保在启动 Notex 之前 Ollama 正在运行
- 更大的模型需要更多的内存和 CPU

### 步骤 3：可选的 Google Gemini（用于信息图）

要使用 Google Gemini Nano Banana 生成信息图：

```env
GOOGLE_API_KEY=your-google-api-key-here
```

从 [https://makersuite.google.com/app/apikey](https://makersuite.google.com/app/apikey) 获取密钥

### 步骤 4：运行应用

配置好 `.env` 文件后，只需运行：

```bash
go run . -server
```

应用将自动加载您的 `.env` 配置并在 `http://localhost:8080` 启动

### 构建并运行（可选）

如果您更喜欢构建二进制文件而不是使用 `go run`：

```bash
go build -o notex .
./notex -server
```

## 📖 使用指南

### 创建笔记本

1. 点击标题栏中的 "新建笔记本" 按钮
2. 输入名称和可选描述
3. 点击 "创建笔记本"

### 添加来源

您可以通过三种方式向笔记本添加内容：

**文件上传**

- 点击 Sources 面板中的 "+" 按钮
- 拖放文件或浏览选择
- 支持格式：PDF、TXT、MD、DOCX、HTML

**粘贴文本**

- 选择 "Text" 标签
- 输入标题并粘贴您的内容

**从 URL**

- 选择 "URL" 标签
- 输入 URL 和可选标题

### 与来源对话

1. 切换到 "CHAT" 标签
2. 向您的内容提问
3. 响应包含相关来源的引用

### 转换功能

点击任意转换卡片即可生成：

| 转换类型    | 描述                                       |
| ----------- | ------------------------------------------ |
| 📝 摘要     | 来源的精简概述                             |
| ❓ FAQ      | 常见问题与答案                             |
| 📚 学习指南 | 包含学习目标的教育材料                     |
| 🗂️ 大纲     | 主题的层次结构                             |
| 🎙️ 播客     | 音频内容的对话脚本                         |
| 📅 时间线   | 来源中的按时间顺序的事件                   |
| 📖 词汇表   | 关键术语和定义                             |
| ✍️ 测验     | 带答案的评估问题                           |
| 📊 信息图   | 内容的手绘风格视觉呈现                     |
| 🧠 思维导图 | 使用 Mermaid.js 生成的来源内容可视化层级图 |

或使用自定义提示字段进行任何其他转换。

### 其他配置选项

对于高级用户，`.env` 文件支持以下额外配置选项：

```env
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 向量存储（默认：sqlite）
# 选项：sqlite、memory、supabase、postgres、redis
VECTOR_STORE_TYPE=sqlite

# RAG 处理
MAX_SOURCES=5          # 检索的最大来源数
CHUNK_SIZE=1000        # 文档分块大小
CHUNK_OVERLAP=200      # 分块重叠

# 文档转换
ENABLE_MARKITDOWN=true  # 使用 Microsoft markitdown 更好地转换 PDF/DOCX

# 播客生成
ENABLE_PODCAST=true
PODCAST_VOICE=alloy    # 选项：alloy、echo、fable、onyx、nova、shimmer

# 功能开关
ALLOW_DELETE=true
ALLOW_MULTIPLE_NOTES_OF_SAME_TYPE=true

# S3 / Ceph 对象存储，文件上传
S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1   # default `us-east-1`
S3_ACCESS_KEY=yourkey
S3_SECRET_KEY=yoursecret
S3_BUCKET=notex-uploads
S3_FORCE_PATH_STYLE=true  # usually true for Ceph/MinIO
S3_SKIP_TLS_VERIFY=true # default false
```

## 🔧 开发

### 运行测试

```bash
go test -v ./...
```

### 构建

```bash
go build -o notex .
```

### 代码质量

```bash
# 格式化
go fmt ./...

# Lint
golangci-lint run

# 检查
go vet ./...
```

### 使用 Makefile

```bash
# 显示所有可用命令
make help

# 开发模式（初始化并运行）
make dev

# 使用 OpenAI 运行
make run-openai

# 使用 Ollama 运行
make run-ollama

# 运行测试
make test

# 代码检查
make check
```

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

## 📄 许可证

Apache License 2.0 - 详见 [LICENSE](./LICENSE)

## 🙏 致谢

- 灵感来自 [Google 的 NotebookLM](https://notebooklm.google.com/)
- 使用 [LangGraphGo](https://github.com/smallnest/langgraphgo) 构建
- 受 [notex](https://github.com/lfnovo/notex) 启发

## 📞 支持

- 在 [GitHub](https://github.com/smallnest/notex/issues) 上报告问题
- 加入 [Notex 社区](https://github.com/smallnest/notex/discussions) 讨论

## 🌟 功能亮点

### 八种智能转换

1. **摘要** - 快速获取文档要点
2. **FAQ** - 自动生成常见问题解答
3. **学习指南** - 创建结构化学习材料
4. **大纲** - 提取内容层次结构
5. **播客** - 生成对话式播客脚本
6. **时间线** - 整理事件的时间顺序
7. **词汇表** - 提取关键术语和定义
8. **测验** - 创建评估问题和答案
9. **信息图** - 生成具有亲和力的手绘风格视觉设计图
10. **思维导图** - 自动生成结构化的知识架构图

### 灵活的知识管理

- 创建多个笔记本组织不同主题
- 混合使用文件、文本和 URL 来源
- 通过 RAG 技术实现智能问答
- 所有转换结果自动保存为笔记

### 隐私优先

- 数据存储在本地 SQLite 数据库
- 可选使用自托管的 PostgreSQL 或 Redis
- 支持 Ollama 进行完全离线的 LLM 推理

---

**Notex** - 注重隐私的开源 NotebookLM 替代方案
https://github.com/smallnest/notex
