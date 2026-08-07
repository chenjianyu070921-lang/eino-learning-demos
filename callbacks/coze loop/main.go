package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/coze-dev/cozeloop-go"
	"github.com/joho/godotenv"
)

func main() {
	cozeloop.SetLogLevel(cozeloop.LogLevelDebug)
	if err := godotenv.Load(); err != nil {
		panic("加载.env文件失败" + err.Error())
	}
	// 1. 准备大模型客户端（图里真正用到的节点）
	ctx := context.Background()
	chatModel := ChatModelClient(ctx)
	// 与 callback 示例不同，这里专注演示“如何把图接入扣子罗盘做分布式追踪”，
	// 所以只用了 chatModel，其余组件（template/embedding/milvus 等）注释掉未用。

	// 2. 初始化“扣子罗盘”（CozeLoop）：一个分布式追踪/可观测性平台。
	//    它会在后台把每个节点的开始/结束、耗时、输入输出上报到云端控制台。
	//    注意：这两行 os.Getenv 只是演示读取，真实的 Token 由 cozeloop.NewClient()
	//    自动从环境变量 COZELOOP_*/ARK_* 读取，无需手动传。
	fmt.Println("model:", os.Getenv("ARK_MODEL_ID"))
	fmt.Println("api:", os.Getenv("ARK_API_KEY"))
	client, err := cozeloop.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close(ctx)
	// NewLoopHandler 把 cozeloop client 包装成 Eino 回调处理器；
	// AppendGlobalHandlers 把它挂成“全局回调”——之后所有图/组件的运行都会自动上报 trace。
	handler := ccb.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(handler)
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
	// ===== 关键修复：读完立刻 Close 流 =====
	output.Close() // ← 加上这行！触发 llm 节点的 OnEndWithStreamOutput 回调，
	//               否则流式结束时 cozeloop 可能收不到最终 trace。

	// 给异步上报留时间（cozeloop 是异步上报，sleep 一下确保 trace 发出去）
	fmt.Println("\n=== 等待 trace 上报 ===")
	time.Sleep(3 * time.Second)

	// 然后再关闭 client
	client.Close(ctx)
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
		case role:
			return "keai", nil
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
