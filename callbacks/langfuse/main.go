package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("加载.env文件失败" + err.Error())
	}
	// 1. 准备大模型客户端（图里真正用到的节点）
	ctx := context.Background()
	chatModel := ChatModelClient(ctx)
	// 与 callback 示例不同，这里专注演示“如何把图接入 Langfuse 做分布式追踪”，
	// 所以只用了 chatModel，其余组件（template/embedding/milvus 等）注释掉未用。

	cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
		Host:      "https://jp.cloud.langfuse.com",
		PublicKey: "pk-lf-769faee7-b581-4425-b901-69cadbff6ce0",
		SecretKey: "sk-lf-59368124-6d4d-4e7c-874c-76bbb1a5772f",
		Name:      "eino-app",
		Release:   "v1.0.0",
	})

	callbacks.AppendGlobalHandlers(cbh)
	//3.编写节点别名

	graph := compose.NewGraph[map[string]string, *schema.Message]()
	graph.AddLambdaNode("lamdba", Lamdba())
	graph.AddLambdaNode("aojiao", Lamdba1())
	graph.AddLambdaNode("keai", Lamdba2())
	graph.AddChatModelNode("llm", chatModel)

	//4.为lamdba节点后面添加branch
	graph.AddBranch("lamdba", Branch())
	//5.搭建边链路
	graph.AddEdge(compose.START, "lamdba")
	graph.AddEdge("aojiao", "llm")
	graph.AddEdge("keai", "llm")
	graph.AddEdge("llm", compose.END)
	//6.开始编译
	compile, err := graph.Compile(ctx)
	if err != nil {
		panic(err)
	}
	output, err := compile.Stream(ctx, map[string]string{
		"role":    "aojiao",
		"content": "我喜欢你",
	}, compose.WithCallbacks(gencallback()))
	if err != nil {
		panic(err)
	}

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
	// 读完流后关闭，释放底层通道
	output.Close()

	// 流式结束回调是异步上报的，等待事件入队后再统一刷新
	fmt.Println("\n=== 等待 trace 上报 ===")
	time.Sleep(3 * time.Second)

	// 刷新并等待所有事件上传完成
	flusher()
}

// 编写节点
func Lamdba() *compose.Lambda {
	lambda := compose.InvokableLambda(func(ctx context.Context, input map[string]string) (output map[string]string, err error) {
		if input["role"] == "aojiao" {
			return map[string]string{"role": "aojiao", "content": input["content"]}, nil
		} else if input["role"] == "keai" {
			return map[string]string{"role": "keai", "content": input["content"]}, nil
		}
		return map[string]string{"role": "aojiao", "content": input["content"]}, nil

	})
	return lambda
}

func Lamdba1() *compose.Lambda {
	lambda := compose.InvokableLambda(func(ctx context.Context, input map[string]string) (output []*schema.Message, err error) {
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个高冷做娇的大小姐，每次都会用傲娇的语气回答我的问题（你的内心是一个病娇）",
			},
			{
				Role:    schema.User,
				Content: input["content"],
			},
		}, nil
	})
	return lambda
}
func Lamdba2() *compose.Lambda {
	lambda := compose.InvokableLambda(func(ctx context.Context, input map[string]string) (output []*schema.Message, err error) {
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个可爱的小女孩，每次都会用可爱的语气回答我的问题",
			},
			{
				Role:    schema.User,
				Content: input["content"],
			},
		}, nil
	})
	return lambda
}
func Branch() *compose.GraphBranch {
	branch := compose.NewGraphBranch(func(ctx context.Context, in map[string]string) (string, error) {
		role := in["role"]
		switch role {
		case "aojiao":
			return "aojiao", nil
		default:
			return "keai", nil
		}
	},
		map[string]bool{
			"aojiao": true,
			"keai":   true},
	)
	return branch
}
func gencallback() callbacks.Handler {
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			fmt.Printf("[trace] %s/%s start\n", info.Component, info.Name)
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			fmt.Printf("[trace] %s/%s end\n", info.Component, info.Name)
			return ctx
		}).Build()
	return handler
}
