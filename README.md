# Eino 全栈学院作业 · 新手速通手册

本仓库是一组围绕 [**Eino**](https://github.com/cloudwego/eino)（字节 CloudWeGo 出品的 Go LLM 应用开发框架）的**独立小示例**。
每个目录都是一个 `go module`（有自己的 `go.mod`），可以单独 `go run .` 跑通，互不干扰，方便逐个知识点学习。

> 适用对象：刚接触 Eino / Go 调用大模型的同学。目标是**看完就能跑、跑完就懂**。

---

## 一、先准备环境（一次搞定）

### 1. 安装 Go
需要 Go 1.25+（示例的 `go.mod` 为 `go 1.25` / `go 1.26.4`）。终端执行 `go version` 验证。

### 2. 准备 `.env`（大模型相关示例都要）
绝大多数示例通过环境变量读取大模型配置。在**要运行的那个目录**下建一个 `.env` 文件：

```dotenv
# 火山方舟（Volcengine Ark）配置
ARK_API_KEY=你的_ark_api_key
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
ARK_MODEL_ID=你的_模型_id（例如 ep-xxxx 或 doubao-xxx）
```

> 需要 `.env` 的目录：
> `chatModel` `chatTemplate` `chain` `graph` `callback` `coze loop` `stateDemo` `RAG flow`
> 其中 `embedding` / `retriever` 用 ARK Embedding，需要 `ARK_API_KEY` + `EMBEDDER`（你的 embedding 模型 id，如 `ep-xxxx`）。
> `indexer` 也会调用 `godotenv.Load()`，目录里需要有 `.env` 文件，但内部连接地址和模型是写死的，文件内容可为空。
> `chatModel` 找不到 `.env` 时不会中断，会退回读取系统环境变量。
> 完全不需要 `.env` 的目录：`eino-tool`（纯工具演示）、`mcp`（完全离线）、`ollama`（直接连本地 Ollama）。

### 3. 按需准备外部服务
| 目录 | 额外依赖 | 说明 |
|------|----------|------|
| `ollama` | 本地安装 [Ollama](https://ollama.com)，并 `ollama pull bge-m3` | 默认 `http://localhost:11434` |
| `indexer` | 本地启动 [Milvus](https://milvus.io)（默认 `127.0.0.1:19530`） | 用 Ollama `bge-m3`，写入集合 `personal_knowledge_base` |
| `retriever` | Milvus + ARK Key | 查集合 `test`，向量维度 1024 |
| `RAG flow` | Milvus + Ollama + ARK Key | 查集合 `test`，检索用 Ollama `bge-m3`，问答用 ARK |
| `coze loop` | ARK Key + 扣子罗盘（CozeLoop）的 Token | 通过 `callbacks.AppendGlobalHandlers` 全局上报 trace |
| `mcp` | 无 | 完全离线，不需要 key |

---

## 二、Eino 核心概念（3 分钟速览）

| 概念 | 一句话 | 对应目录 |
|------|--------|----------|
| **ChatModel** | 调大模型聊天/生成文本（支持流式） | `chatModel` `graph` `chain` |
| **ChatTemplate** | 用 `{变量}` 占位符拼 prompt 模板 | `chatTemplate` `chain` |
| **Embedding** | 把文本变成向量（用于检索/相似度） | `embedding` `ollama` `retriever` |
| **Tool** | 给大模型挂载「外部能力」（函数调用） | `eino-tool` `graph` |
| **Graph / Chain** | 把上面这些节点用「边」连成流程图 | `chain` `graph` `stateDemo` `RAG flow` |
| **State（图状态）** | 图运行期间所有节点共享读写的「白板」 | `graph/stateDemo.go` `callback` `stateDemo` |
| **Branch（分支）** | 按条件把数据路由到不同节点 | `graph/stateDemo.go` `callback` `coze loop` |
| **Indexer / Retriever** | 往向量库「写」/ 从向量库「查」 | `indexer` `retriever` `RAG flow` |
| **Callback** | 监听节点开始/结束，做日志、追踪 | `callback` `coze loop` |
| **MCP** | 模型上下文协议，定义「客户端如何发现并调用能力」 | `mcp` |

> **边 vs State 的区别**（新手常混）：
> - **边（Edge）**：数据像水管一样，从 A 节点的输出流向 B 节点的输入，类型必须一一对应。
> - **State**：一张在整张图运行期间一直存活的共享对象，任何节点都能读写，但**不靠边传递**。适合存「对话历史、跨节点共享的中间结果」。

---

## 三、目录导航 & 运行命令

下面每个目录都可 `cd` 进去后直接 `go run .`（需要 `.env` 的先放好 `.env`）。

### 🟢 入门级（不依赖外部服务 / 只调一个模型）

- **`chatModel/`** — 最基础：创建 ChatModel 并**流式对话**。支持 `--instruction` 自定义系统提示词，问题写在命令后面。
  ```bash
  cd chatModel && go run . --instruction "你是个go领域的专家" "什么是Eino"
  ```
- **`chatTemplate/`** — 用 `{task}` 占位符渲染 prompt 模板，再喂给 ChatModel。
  ```bash
  cd chatTemplate && go run .
  ```
- **`embedding/`** — 用火山方舟多模态 Embedding 把文本变成向量，并打印向量维度。
  ```bash
  cd embedding && go run .
  ```
- **`ollama/`** — 直接调 Ollama 官方 Go SDK 的 `bge-m3` 做 embedding（**需先装 Ollama**，不经 Eino 封装）。
  ```bash
  cd ollama && go run .
  ```
- **`eino-tool/`** — 定义一个「查大模型官网链接」的工具，并手动调用它。
  ```bash
  cd eino-tool && go run .
  ```
- **`mcp/`** — MCP 服务端 + 客户端示例（Tools / Resources / Prompts 三原语），**完全离线**。
  ```bash
  cd mcp && go run ./client     # 客户端会自动拉起服务端演示
  ```

### 🟡 进阶级（组合多个组件）

- **`chain/`** — 用 `compose.NewChain` 把「模板 + 聊天模型」串成直线（`newEmbedding.go` 是预留的 embedding 帮助函数，本例未用到）。
  ```bash
  cd chain && go run .
  ```
- **`graph/`** — 重点目录：
  - `main.go`：**ReAct Agent**，给 ChatModel 挂载 Tool，让模型自己决定调工具（流式输出）。
    ```bash
    cd graph && go run .
    ```
  - `stateDemo.go`：**State + Branch 综合示例**（傲娇/可爱两条分支，用 State 存历史）。⚠️ 该文件和 `main.go` 同属 `package main`，不能一起 `go run .`，要单独指定文件运行：
    ```bash
    cd graph && go run stateDemo.go newChatModel.go
    ```
- **`stateDemo/`** — 独立的 State+Branch 示例（自带 `go.mod`，和上面 `graph/stateDemo.go` 思路相同但可单独 `go run .`）。
  ```bash
  cd stateDemo && go run .
  ```
- **`callback/`** — 演示 `State` 的读写（`ProcessState`）和 Callback 追踪：用 `callbacks.NewHandlerBuilder()` + `compose.WithCallbacks()` 打印每个节点开始/结束（`cozeloop_trace2.png` 是扣子罗盘 trace 界面截图，供参考）。
  ```bash
  cd callback && go run .
  ```

### 🔴 高级 / 依赖外部服务

- **`indexer/`** — 把 markdown 知识库切分、向量化并写入本地 Milvus（**需先启动 Milvus 和 Ollama**；入口文件名为 `mian.go`，属历史拼写，不影响 `go run .`）。
  ```bash
  cd indexer && go run .
  ```
- **`retriever/`** — 从 Milvus 按问题检索相关文档（**需 Milvus + ARK Key**）。
  ```bash
  cd retriever && go run .
  ```
- **`RAG flow/`** — 完整 RAG：先让模型判断要不要检索（意图路由），需要时检索 → 拼 prompt → 大模型流式作答。
  ```bash
  cd "RAG flow" && go run .
  ```
- **`coze loop/`** — 给图接入「扣子罗盘」分布式追踪（**需 CozeLoop Token**，同时图里用到了 ARK 模型）。
  ```bash
  cd "coze loop" && go run .
  ```

---

## 四、给初学者的学习路线（建议顺序）

1. `chatModel` → 先搞懂「怎么调大模型、流式是什么」。
2. `chatTemplate` → 搞懂「prompt 模板怎么拼」。
3. `eino-tool` → 搞懂「工具（函数调用）长什么样」。
4. `chain` → 用 `NewChain` 把「模板 + 模型」连成直线，理解节点类型对齐。
5. `graph/stateDemo.go`（或 `stateDemo/`）→ 理解 **State** 与 **Branch**。
6. `graph/main.go` → 理解 **ReAct Agent**：模型自己决定调哪个 Tool。
7. `embedding` / `ollama` → 理解向量化。
8. `indexer` / `retriever` / `RAG flow` → 理解 RAG 全链路（注意各示例的集合名差异）。
9. `callback` / `coze loop` → 理解可观测性（日志/追踪）。
10. `mcp` → 理解更通用的「能力暴露协议」。

---

## 五、常见问题（FAQ）

**Q1：`panic: ... output type [...] and end node's input type [...] mismatch`**
节点之间用「边」连接时，上游输出类型必须等于下游输入类型。检查 `AddEdge` 两端节点的函数签名（`map[string]string` vs `map[string]any` 等）。

**Q2：`panic: branch end node 'xxx' needs to be added to graph first`**
`AddBranch` 调用时，分支指向的目标节点必须**先**用 `AddLambdaNode`/`AddChatModelNode` 加入图。把 `AddBranch` 放到所有目标节点添加之后。

**Q3：`graph/stateDemo.go` 直接 `go run .` 报 `main redeclared`？**
`graph/` 目录下已经有 `main.go`（ReAct Agent），两个 `main` 冲突。请用 `go run stateDemo.go newChatModel.go` 单独跑这份。

**Q4：Milvus 相关示例连不上 / 查不到结果？**
确认本地 Milvus 已启动在 `127.0.0.1:19530`，且**集合名对得上**：`indexer` 写入 `personal_knowledge_base`，而 `retriever` / `RAG flow` 默认查 `test`。想串起来跑，把两边集合名改成一致（改 `newRetriever.go` 里的 `Collection`，或改 `indexer` 里的集合名）。向量维度统一为 `bge-m3` 的 1024 维。

**Q5：报错 `dial tcp 127.0.0.1:11434`？**
`ollama` / `indexer` / `RAG flow` 需要本地装好 Ollama 并拉取 `bge-m3` 模型。

**Q6：想看每个节点的执行日志？**
参考 `callback/` 或 `coze loop/`，用 `compose.WithCallbacks(handler)` 给图挂上回调；`coze loop` 展示的是把回调接进分布式追踪平台。

**Q7：目录里的 `.exe` 文件是什么？**
之前 `go build` 生成的编译产物（如 `chatModel/ch01.exe`、`graph/graph.exe`），可以直接忽略或删除，不影响 `go run .`。

---

## 六、一个最小的「图」长什么样（伪代码）

```go
g := compose.NewGraph[string, string]()      // [输入类型, 输出类型]
g.AddLambdaNode("node1", compose.InvokableLambda(fn1))
g.AddLambdaNode("node2", compose.InvokableLambda(fn2))
g.AddEdge(compose.START, "node1")            // 起点 → node1
g.AddEdge("node1", "node2")                 // node1 → node2
g.AddEdge("node2", compose.END)             // node2 → 终点
r, _ := g.Compile(ctx)
out, _ := r.Invoke(ctx, "你好")
```

记住套路：**`compose.START` 接第一个节点，`compose.END` 收最后一个节点，中间用 `AddEdge` 或 `AddBranch` 连。**

---
祝学习顺利 🚀 遇到编译/运行问题，先对照第五节 FAQ，再回看对应目录代码里的注释。
