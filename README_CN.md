# Notex - Open Notebook

<div align="center">

**注重隐私的开源 NotebookLM 替代方案**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](./LICENSE)

AI 驱动知识管理应用程序，让您从文档中创建智能笔记本。

**项目地址：** https://github.com/smallnest/notex

</div>

## ✨ 特性

- 📚 **多种来源类型** - 支持上传 PDF、文本文件、Markdown、DOCX 和 HTML 文档
- 🤖 **AI 驱动对话** - 基于您的来源提问并获得答案
- ✨ **多种转换** - 生成摘要、FAQ、学习指南、大纲、时间线、词汇表、测验和播客脚本
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

# 设置您的 API 密钥
export OPENAI_API_KEY=your_key_here

# 运行服务器
go run . -server
```

在浏览器中打开 `http://localhost:8080`

### 使用 Ollama（本地、免费）

```bash
# 确保 Ollama 正在运行
ollama serve

# 使用 Ollama 运行
export OLLAMA_BASE_URL=http://localhost:11434
go run . -server
```

### 或者：构建后运行

```bash
# 构建二进制文件
go build -o open-notebook .

# 使用 OpenAI 运行
export OPENAI_API_KEY=your_key_here
./open-notebook -server

# 或使用 Ollama 运行
export OLLAMA_BASE_URL=http://localhost:11434
./open-notebook -server
```

## 📖 使用指南

### 创建笔记本

1. 点击标题栏中的 "New Notebook" 按钮
2. 输入名称和可选描述
3. 点击 "Create Notebook"

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

| 转换类型 | 描述 |
|---------|------|
| 📝 摘要 | 来源的精简概述 |
| ❓ FAQ | 常见问题与答案 |
| 📚 学习指南 | 包含学习目标的教育材料 |
| 🗂️ 大纲 | 主题的层次结构 |
| 🎙️ 播客 | 音频内容的对话脚本 |
| 📅 时间线 | 来源中的按时间顺序的事件 |
| 📖 词汇表 | 关键术语和定义 |
| ✍️ 测验 | 带答案的评估问题 |

或使用自定义提示字段进行任何其他转换。

## ⚙️ 配置

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `OPENAI_API_KEY` | OpenAI API 密钥 | 必需（除非使用 Ollama）|
| `OPENAI_BASE_URL` | 自定义 API 基础 URL | OpenAI 默认值 |
| `OPENAI_MODEL` | 模型名称 | `gpt-4o-mini` |
| `EMBEDDING_MODEL` | 嵌入模型 | `text-embedding-3-small` |
| `OLLAMA_BASE_URL` | Ollama 服务器 URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | Ollama 模型名称 | `llama3.2` |
| `SERVER_HOST` | 服务器主机 | `0.0.0.0` |
| `SERVER_PORT` | 服务器端口 | `8080` |
| `VECTOR_STORE_TYPE` | 向量存储后端 | `sqlite` |
| `STORE_PATH` | 数据库路径 | `./data/checkpoints.db` |
| `MAX_SOURCES` | RAG 的最大来源数 | `5` |
| `CHUNK_SIZE` | 文档分块大小 | `1000` |
| `CHUNK_OVERLAP` | 分块重叠 | `200` |

### 向量存储选项

- `sqlite` - 本地 SQLite 数据库（默认）
- `memory` - 内存向量
- `supabase` - Supabase 向量存储
- `postgres` / `pgvector` - 带 pgvector 的 PostgreSQL
- `redis` - 带 RediSearch 的 Redis

### 配置文件示例

**docker-compose.yml**（用于 PostgreSQL + pgvector）

```yaml
version: '3.8'
services:
  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_DB: notebook
      POSTGRES_USER: notebook
      POSTGRES_PASSWORD: secret
    ports:
      - "5432:5432"

  app:
    build: .
    environment:
      - POSTGRES_URL=postgres://notebook:secret@postgres:5432/notebook
      - VECTOR_STORE_TYPE=postgres
    ports:
      - "8080:8080"
```

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端                                  │
│              (HTML/CSS/JS - 野兽派 UI)                       │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTP API
┌────────────────────▼────────────────────────────────────────┐
│                      服务器层                                │
│                    (Gin 路由器)                              │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
┌───────▼──────┐ ┌──▼──────┐ ┌──▼──────────┐
│  VectorStore │ │  Store  │ │   Agent     │
│              │ │         │ │             │
│ - 嵌入       │ │ SQLite  │ │ - LLM 调用  │
│ - 搜索       │ │         │ │ - 提示工程  │
│ - 分块       │ │         │ │ - RAG       │
└──────────────┘ └─────────┘ └─────────────┘
```

## 🎨 设计理念

本应用采用 **学术野兽派** 美学风格：

- **暖色调纸张** 配以锐利的黑色墨水 - 如同档案文件
- **等宽技术字体** 搭配优雅的衬线标题
- **可见的网格结构** - 展示"知识结构"
- **高对比度排版** 以提高可读性和专注度
- **细微的纹理** 增加温暖感和深度

设计强调功能胜于形式，使内容成为主角，同时保持独特、难忘的特色。

## 📁 项目结构

```
notex/
├── backend/
│   ├── main.go          # CLI 入口点
│   ├── config.go        # 配置管理
│   ├── types.go         # 数据结构
│   ├── store.go         # 数据库持久化
│   ├── vector.go        # 向量搜索
│   ├── agent.go         # AI 操作
│   └── server.go        # HTTP 服务器
├── frontend/
│   ├── index.html       # 主 HTML
│   └── static/
│       ├── style.css    # 野兽派样式
│       └── app.js       # 应用逻辑
├── go.mod
├── go.sum
├── main.go
├── Dockerfile
├── Makefile
├── docker-compose.yml
└── README.md
```

## 🔧 开发

### 运行测试

```bash
go test -v ./...
```

### 构建

```bash
go build -o open-notebook .
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

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务（PostgreSQL + Redis + 应用）
docker-compose up -d

# 查看日志
docker-compose logs -f app

# 停止服务
docker-compose down
```

### 单独构建

```bash
# 构建镜像
docker build -t open-notebook .

# 运行容器
docker run -p 8080:8080 \
  -e OPENAI_API_KEY=your_key \
  -v $(pwd)/data:/app/data \
  open-notebook
```

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

## 📄 许可证

Apache License 2.0 - 详见 [LICENSE](./LICENSE)

## 🙏 致谢

- 灵感来自 [Google 的 NotebookLM](https://notebooklm.google.com/)
- 使用 [LangGraphGo](https://github.com/smallnest/langgraphgo) 构建
- 由 [LangChain Go](https://github.com/tmc/langchaingo) 提供支持

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
