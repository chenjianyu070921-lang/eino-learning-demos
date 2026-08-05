// MCP 客户端示例：启动服务端子进程，然后依次演示握手、列工具、调工具、
// 读资源、取提示词。
// 运行方式：go run ./client
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. 创建客户端：通过 stdio 启动服务端子进程
	//    参数依次是：可执行命令、环境变量、命令参数
	c, err := client.NewStdioMCPClient("go", os.Environ(), "run", ".")
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer c.Close()

	// 2. 初始化握手：声明协议版本和客户端信息
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "go-mcp-demo-client",
		Version: "0.1.0",
	}

	initResult, err := c.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	fmt.Printf("已连接服务端: %s (version %s)\n\n",
		initResult.ServerInfo.Name,
		initResult.ServerInfo.Version,
	)

	// 3. 依次演示 MCP 的三种能力
	listTools(ctx, c)
	callAdd(ctx, c)
	callGreet(ctx, c)
	readWelcome(ctx, c)
	getGreetPrompt(ctx, c)
}

// listTools 列出服务端注册的所有工具及其参数结构
func listTools(ctx context.Context, c *client.Client) {
	fmt.Println("===== 1. 列出工具 (tools/list) =====")
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("列出工具失败: %v", err)
	}

	for _, tool := range tools.Tools {
		fmt.Printf("工具名: %s\n", tool.Name)
		fmt.Printf("描述: %s\n", tool.Description)
		schema, _ := json.MarshalIndent(tool.InputSchema, "  ", "  ")
		fmt.Printf("参数结构:\n  %s\n", string(schema))
	}
	fmt.Println()
}

// callAdd 调用 add 工具
func callAdd(ctx context.Context, c *client.Client) {
	fmt.Println("===== 2. 调用工具 (tools/call): add =====")
	req := mcp.CallToolRequest{}
	req.Params.Name = "add"
	req.Params.Arguments = map[string]any{
		"a": 1,
		"b": 2,
	}

	result, err := c.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("调用 add 失败: %v", err)
	}
	printToolResult(result)
}

// callGreet 调用 greet 工具（演示布尔参数）
func callGreet(ctx context.Context, c *client.Client) {
	fmt.Println("===== 3. 调用工具 (tools/call): greet =====")
	req := mcp.CallToolRequest{}
	req.Params.Name = "greet"
	req.Params.Arguments = map[string]any{
		"name":   "小明",
		"formal": true,
	}

	result, err := c.CallTool(ctx, req)
	if err != nil {
		log.Fatalf("调用 greet 失败: %v", err)
	}
	printToolResult(result)
}

// readWelcome 读取服务端提供的资源
func readWelcome(ctx context.Context, c *client.Client) {
	fmt.Println("===== 4. 读取资源 (resources/read) =====")
	req := mcp.ReadResourceRequest{}
	req.Params.URI = "demo://welcome"

	result, err := c.ReadResource(ctx, req)
	if err != nil {
		log.Fatalf("读取资源失败: %v", err)
	}
	for _, content := range result.Contents {
		if text, ok := content.(mcp.TextResourceContents); ok {
			fmt.Printf("内容: %s\n", text.Text)
		}
	}
	fmt.Println()
}

// getGreetPrompt 获取提示词模板并填入参数
func getGreetPrompt(ctx context.Context, c *client.Client) {
	fmt.Println("===== 5. 获取提示词 (prompts/get) =====")
	req := mcp.GetPromptRequest{}
	req.Params.Name = "greet_prompt"
	req.Params.Arguments = map[string]string{"name": "小红"}

	result, err := c.GetPrompt(ctx, req)
	if err != nil {
		log.Fatalf("获取提示词失败: %v", err)
	}
	for _, msg := range result.Messages {
		if text, ok := msg.Content.(mcp.TextContent); ok {
			fmt.Printf("[%s] %s\n", msg.Role, text.Text)
		}
	}
}

// printToolResult 打印工具调用结果
func printToolResult(result *mcp.CallToolResult) {
	for _, content := range result.Content {
		if text, ok := content.(mcp.TextContent); ok {
			fmt.Printf("结果: %s\n\n", text.Text)
			return
		}
	}
	fmt.Printf("结果(非文本): %+v\n\n", result.Content)
}
