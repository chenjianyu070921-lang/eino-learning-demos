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
	// ark 普通 ChatModel 仅支持普通 Message，直接调用 runNormal 跑流式对话
	runNormal(ctx, instruction, query)

}
func runNormal(ctx context.Context, instruction string, query string) {
	// 创建 ChatModel 客户端：从环境变量读取 APIKey / 模型ID / 接入地址。
	// Thinking 设为 Disabled 表示关闭“深度思考”模式（有些模型支持思考过程）。
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

	// 构造消息列表：System 设定角色/口吻，User 是用户真正的问题。
	// 注意类型固定是 []*schema.Message 切片，和 Stream 的入参一致。
	message := []*schema.Message{
		schema.SystemMessage(instruction), //系统消息（人设/指令）
		schema.UserMessage(query),         //用户提问消息
	}

	// Stream 返回“流”，模型每生成一小段就立刻吐出来，实现“打字机”效果。
	_, _ = fmt.Fprint(os.Stdout, "[assistant] ")
	stream, err := model.Stream(ctx, message)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer stream.Close() // 读完一定要 Close，释放连接

	// 循环从流里取数据：io.EOF 表示流结束；其余错误才中断。
	for {
		recv, err := stream.Recv() //流式输出：每次取一小段（chunk）
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if recv != nil {
			_, _ = fmt.Fprint(os.Stdout, recv.Content) // 追加打印这一段（不换行）
		}
	}
	_, _ = fmt.Fprintln(os.Stdout) // 流结束后补一个换行
}
