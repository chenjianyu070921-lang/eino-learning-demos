package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("加载.env文件失败：" + err.Error())
	}
	fmt.Println(".env加载成功")

	ctx := context.Background()
	chatModel := ChatModelClient(ctx)
	invokableTool := CreateTool()

	info, err := invokableTool.Info(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("工具名称: %s\n工具描述: %s\n", info.Name, info.Desc)

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		Model: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{invokableTool},
		},
		MessageModifier: react.NewPersonaModifier(
			"你是网站助手。用户询问大模型或游戏的官方网址时，必须调用 get_llm_url 工具查询，不能编造链接；取得工具结果后再简洁回答。",
		),
		MaxStep: 8, //防止模型无线循环调用
	})
	if err != nil {
		panic(err)
	}

	messages := []*schema.Message{
		schema.UserMessage("请告诉我 原神 的官方网址，并简单介绍这个网站。"),
	}
	output, err := agent.Stream(ctx, messages)
	if err != nil {
		panic(err)
	}
	defer output.Close()
	for {
		// 从流式流中接收一段模型返回的数据分片（chunk）
		recv, err := output.Recv()
		// 判断是否读到流末尾：io.EOF 代表流式输出全部接收完成
		if errors.Is(err, io.EOF) {
			break
		}
		// 出现非EOF错误，代表流式传输异常
		if err != nil {
			// 将错误信息打印到标准错误输出
			_, _ = fmt.Fprintln(os.Stderr, err)
			// 异常退出程序
			os.Exit(1)
		}
		// 判断当前分片不为空，避免空数据打印
		if recv != nil {
			// 将模型返回的文本分片持续打印到标准输出（不换行，实现打字机流式效果）
			_, _ = fmt.Fprint(os.Stdout, recv.Content)
		}
	}
	// 流式接收循环结束后，输出换行，把命令行光标换到下一行
	_, _ = fmt.Fprintln(os.Stdout)
}


