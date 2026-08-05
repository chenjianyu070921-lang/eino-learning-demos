# Go MCP 简易 Demo

这是一个用 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) 写的 MCP（Model Context Protocol）入门示例，包含一个服务端和一个客户端，覆盖 MCP 的三个核心原语：

- **Tools（工具）**：客户端按名字调用，服务端执行并返回结果。
- **Resources（资源）**：客户端按 URI 读取服务端提供的数据。
- **Prompts（提示词模板）**：客户端获取一段可填入参数的提示词消息。

## 目录结构

```text
mcp/
├── main.go       # MCP 服务端：注册 2 个工具、1 个资源、1 个提示词
├── client/
│   └── main.go   # MCP 客户端：启动服务端并调用它
├── go.mod
└── README.md
```

## 运行

```bash
cd C:\Users\crema\Desktop\全栈学院作业\eino\mcp
go mod tidy
go run ./client
```

客户端会通过标准输入输出（stdio）自动启动根目录的 `main.go` 作为子进程，依次演示：

1. 初始化握手（initialize）
2. 列出工具（tools/list）
3. 调用 `add` 工具（tools/call）
4. 调用 `greet` 工具（tools/call）
5. 读取 `demo://welcome` 资源（resources/read）
6. 获取 `greet_prompt` 提示词（prompts/get）

单独运行服务端：

```bash
go run .
```

stdio 服务端不会自己输出内容，它会一直等待客户端接入。这属于正常现象，按 `Ctrl+C` 即可退出。

## 在 MCP 客户端中接入

也可以把 `server` 当做一个真正的 MCP Server，配置到 Claude Desktop、Cursor 等支持 MCP 的软件里。

先编译出可执行文件：

```bash
go build -o mcp-server.exe .
```

然后在 MCP 客户端里添加一个 stdio Server，命令填 `mcp-server.exe` 的完整路径即可。

## 学习建议

- 先看根目录的 `main.go`，理解服务端如何注册 Tool / Resource / Prompt。
- 再看 `client/main.go`，理解客户端如何发现和调用这些能力。
- 修改 `main.go` 里的工具，再运行 `go run ./client`，观察列表和调用结果的变化。

这个 demo 不需要 API Key，也不访问网络；MCP 本身只是定义“客户端怎么发现和调用能力”的协议，真正的业务逻辑（如调用大模型、查数据库）都写在服务端里。
