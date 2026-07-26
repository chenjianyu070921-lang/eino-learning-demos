package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 中的密钥（目录里没有 .env 时会尝试用系统环境变量，仅提示不中断）
	if err := godotenv.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "提示: 未找到 .env 文件，将尝试使用系统环境变量")
	}

	//用来存放系统提示词
	var instruction string
	//绑定提示词
	flag.StringVar(&instruction, "instruction", "你是个go领域的专家", "")
	flag.Parse()
	//提取用户的问题（flag 之后的位置参数）
	query := strings.TrimSpace(strings.Join(flag.Args(), " "))
	//如果用户没写问题，返回用法提示
	if query == "" {
		_, _ = fmt.Fprintln(os.Stderr, `usage: go run . --instruction "系统提示词" "你的问题"`)
		os.Exit(2)
	}

	ctx := context.Background()
	// ark普通ChatModel仅支持普通Message，直接调用runNormal
	runNormal(ctx, instruction, query)

}
func runNormal(ctx context.Context, instruction string, query string) {
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey:  os.Getenv("ARK_API_KEY"),
		Model:   os.Getenv("ARK_MODEL_ID"),
		BaseURL: os.Getenv("ARK_BASE_URL"),
		Thinking: &arkModel.Thinking{Type: arkModel.ThinkingTypeDisabled},
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 固定切片类型 []*schema.Message，和 Stream 入参匹配
	message := []*schema.Message{
		schema.SystemMessage(instruction), //系统消息
		schema.UserMessage(query),         //用户提问消息
	}

	_, _ = fmt.Fprint(os.Stdout, "[assistant] ")
	stream, err := model.Stream(ctx, message)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer stream.Close()

	for {
		recv, err := stream.Recv() //流式输出
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if recv != nil {
			_, _ = fmt.Fprint(os.Stdout, recv.Content)
		}
	}
	_, _ = fmt.Fprintln(os.Stdout)
}
