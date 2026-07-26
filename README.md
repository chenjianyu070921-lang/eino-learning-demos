# Eino 学习实例集

基于字节跳动 **Eino**（Go 语言的 LLM 应用开发框架）的一系列**新手友好型**示例代码，覆盖从最基础的「调一次模型对话」到完整的「RAG 检索增强生成」全链路。

> 每个子目录都是一个**独立可运行的 Go module**，可以单独 `cd` 进去 `go run .` 跑起来学习，互不干扰。

---

## 一、环境准备

| 依赖 | 用途 | 安装/启动 |
|------|------|-----------|
| **Go** ≥ 1.21 | 运行所有示例 | [官网下载](https://go.dev/dl/) |
| **Ollama** | 本地 Embedding（`bge-m3`），`graph`/`indexer`/`chain` 用到 | `ollama serve` + `ollama pull bge-m3` |
| **Milvus** | 向量数据库（RAG 用） | 本地 standalone 版，监听 `127.0.0.1:19530` |
| **火山方舟 ARK Key** | 对话 / 多模态 Embedding | 在 [火山方舟控制台](https://console.volcengine.com/ark) 开通，拿到 `ARK_API_KEY` 与模型接入点 `ARK_MODEL_ID` |

---

## 二、配置密钥（重要）

所有敏感信息放在各目录的 `.env` 文件里，这些文件**已被 `.gitignore` 忽略，不会提交到 git**（模板见仓库根目录的 `.env.example`）。

首次使用，把模板复制成你需要的 `.env`：

```bash
cp .env.example graph/.env
cp .env.example indexer/.env
cp .env.example chain/.env
cp .env.example embedding/.env
cp .env.example "RAG flow/.env"
```

然后编辑各 `.env`，填入你真实的 `ARK_API_KEY`、`ARK_MODEL_ID` 等。

> ⚠️ 不要把真实 `.env` 提交到任何公开仓库。

---

## 三、目录结构与学习顺序

| 目录 | 演示内容 | 依赖 |
|------|----------|------|
| `chatModel/` | 最基础：用 ARK 模型做**流式对话** | ARK Key |
| `chatTemplate/` | 用 `ChatTemplate` 渲染提示词 | ARK Key |
| `embedding/` | 用 ARK **多模态 Embedding** 把文本变向量 | ARK Key |
| `ollama/` | 直接调 Ollama 本地 embedding | Ollama |
| `eino-tool/` | 定义一个**工具（Tool）**并调用 | 无 |
| `chain/` | 用 `Chain` 把「模板 + 模型」串成直线 | Ollama + ARK |
| `graph/` | 用 `Graph` + **React Agent** 做「工具调用」编排 | ARK Key |
| `indexer/` | 把 markdown 知识库切分、向量化、**写入** Milvus | Ollama + Milvus |
| `retriever/` | 从 Milvus **检索**最相关的资料 | ARK Key + Milvus |
| `RAG flow/` | 手写「意图路由」的 RAG（查不查库自定） | Ollama + Milvus + ARK |

**推荐学习路线**：
`chatModel` → `chatTemplate` → `embedding` → `eino-tool` → `chain` → `graph` → `indexer` → `retriever` → `RAG flow`

---

## 四、跑一个完整 RAG 链路

1. 启动 Ollama 与 Milvus；
2. 准备知识库文件 `个人成长与求职知识库.md`（放在 `indexer/` 目录下）；
3. **写入向量库**：
   ```bash
   cd indexer
   go mod tidy
   go run .
   ```
   会在 Milvus 的 `personal_knowledge_base` 集合里建表并存入向量（含 `content` + `metadata`）。
4. **问答**：
   ```bash
   cd graph
   go mod tidy
   go run .
   ```
   React Agent 会调用 `get_llm_url` 等工具，并结合资料回答。

---

## 五、新手易踩的坑

- **Embedding 维度要一致**：写入（`indexer` 用 Ollama `bge-m3` = 1024 维）和检索（`retriever`/`graph` 用同一模型）必须**同源同维度**，否则 Milvus 报 `vector dimension mismatch`。
- **集合名要对上**：`indexer` 写入 `personal_knowledge_base`；`graph` 的检索也查此集合；`RAG flow` 示例默认查 `test` 集合，按需改 `RAG flow/newRetriever.go` 里的 `Collection`。
- **`.env` 别提交**：已在 `.gitignore` 中忽略；改集合名、TopK、地址等直接编辑对应的 `newRetriever.go` / `newIndexer.go`。
- **每个目录是独立 module**：首次在某个目录运行前，记得先 `go mod tidy` 拉依赖。

---

## 六、许可证

仅用于学习交流。
