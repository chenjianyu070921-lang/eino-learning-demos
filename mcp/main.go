// MCP 服务端示例：定义工具、资源和提示词，并通过 stdio 提供服务。
// 运行方式：go run .
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 1. 创建 MCP Server（名称和版本会显示给客户端）
	s := server.NewMCPServer(
		"go-mcp-demo-server",
		"0.1.0",
	)

	// 2. 注册 Tools（工具）：LLM 或客户端可以按名称调用
	s.AddTool(mcp.NewTool("add",
		mcp.WithDescription("计算两个数字的和"),
		mcp.WithNumber("a", mcp.Description("第一个数字"), mcp.Required()),
		mcp.WithNumber("b", mcp.Description("第二个数字"), mcp.Required()),
	), handleAdd)

	s.AddTool(mcp.NewTool("greet",
		mcp.WithDescription("根据名字生成一句问候语"),
		mcp.WithString("name", mcp.Description("你的名字"), mcp.Required()),
		mcp.WithBoolean("formal", mcp.Description("是否使用正式语气")),
	), handleGreet)

	// 3. 注册 Resources（资源）：给客户端读取的静态/动态数据
	s.AddResource(mcp.NewResource("demo://welcome",
		"欢迎信息",
		mcp.WithMIMEType("text/plain"),
		mcp.WithResourceDescription("一段给客户端的欢迎文本"),
	), handleReadWelcome)

	// 4. 注册 Prompts（提示词模板）：客户端拿去组装给 LLM 的消息
	s.AddPrompt(mcp.NewPrompt("greet_prompt",
		mcp.WithPromptDescription("生成一个打招呼的提示词"),
		mcp.WithArgument("name", mcp.ArgumentDescription("对方的名字"), mcp.RequiredArgument()),
	), handleGreetPrompt)

	// 5. 通过标准输入输出启动服务
	//    stdio 是 MCP 最常见的本地传输方式（Claude Desktop、Cursor 等都支持）
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

// 工具处理函数的签名是固定的：接收请求，返回结果或错误
func handleAdd(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 参数是 map[string]any，JSON 里的数字会解析成 float64
	args := req.GetArguments()
	a, _ := args["a"].(float64)
	b, _ := args["b"].(float64)

	return mcp.NewToolResultText(fmt.Sprintf("%v + %v = %v", a, b, a+b)), nil
}

func handleGreet(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name, _ := args["name"].(string)
	formal, _ := args["formal"].(bool)

	if formal {
		return mcp.NewToolResultText(fmt.Sprintf("尊敬的 %s，您好！很高兴为您服务。", name)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("你好，%s！欢迎学习 MCP。", name)), nil
}

// 资源处理函数：返回资源内容切片
func handleReadWelcome(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "demo://welcome",
			MIMEType: "text/plain",
			Text:     "欢迎使用 Go MCP Demo！当前时间：" + time.Now().Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// 提示词处理函数：返回组装好的消息列表
func handleGreetPrompt(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	name := "朋友"
	if v := req.Params.Arguments["name"]; v != "" {
		name = v
	}

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(fmt.Sprintf("请用热情友好的语气向 %s 打招呼。", name))),
	}
	return mcp.NewGetPromptResult("给用户打招呼的提示词", messages), nil
}
