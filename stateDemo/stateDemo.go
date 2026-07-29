package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

type State struct {
	History map[string]string
}

func NewState(ctx context.Context) *State {
	return &State{History: make(map[string]string)}
}
func main() {
	if err := godotenv.Load(); err != nil {
		panic("加载.env文件失败：" + err.Error())
	}
	fmt.Println(".env加载成功")
	ctx := context.Background()

	graph := compose.NewGraph[map[string]string, *schema.Message](
		compose.WithGenLocalState(NewState),
	)

	lamdba := compose.InvokableLambda(func(ctx context.Context,
		input map[string]string) (output map[string]string, err error) {
		//在节点内部处理state
		_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
			state.History["aojiao_action"] = "我喜欢你"
			state.History["keai_action"] = "摸摸头"
			return nil
		})
		if input["role"] == "aojiao" {
			return map[string]string{"role": "aojiao", "content": input["content"]}, nil
		}
		if input["role"] == "keai" {
			return map[string]string{"role": "keai", "content": input["content"]}, nil
		}
		return map[string]string{"role": "user", "content": input["content"]}, nil
	})
	//建立分支
	aojiaoLamdba := compose.InvokableLambda(func(ctx context.Context,
		input map[string]string) (output []*schema.Message, err error) {
		//在节点内部处理state
		_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
			input["content"] = input["content"] + state.History["aojiao_action"]
			return nil
		})
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个高冷做娇的大小姐，每次都会用傲娇的语气回答我的问题（你的内心是一个病娇）",
			},
			{
				Role:    schema.User,
				Content: input["content"],
			},
		}, err
	})

	keailambda := compose.InvokableLambda(func(ctx context.Context,
		input map[string]string) (output []*schema.Message, err error) {
		_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
			input["content"] = input["content"] + state.History["keai_action"]
			return nil
		})
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个可爱的小女孩，每次都会用可爱的语气回答我的问题",
			},
			{
				Role:    schema.User,
				Content: input["content"],
			},
		}, err
	})
	branch := compose.NewGraphBranch(
		func(ctx context.Context, in map[string]string) (string, error) {
			role := in["role"]
			switch role {
			case "aojiao":
				return "lambda1", nil
			case "keai":
				return "lambda2", nil
			default:
				return "lambda1", nil
			}
		},
		// 第二个参数：所有分支可跳转节点列表
		map[string]bool{
			"lambda1": true,
			"lambda2": true,
		},
	)

	//编写节点

	err := graph.AddLambdaNode("lambda", lamdba)
	if err != nil {
		panic(err)
	}
	err = graph.AddLambdaNode("lambda1", aojiaoLamdba)
	if err != nil {
		panic(err)
	}
	err = graph.AddLambdaNode("lambda2", keailambda)
	if err != nil {
		panic(err)
	}
	err = graph.AddChatModelNode("llm", ChatModelClient(ctx))
	if err != nil {
		panic(err)
	}

	//
	// 2. 绑定分支：lambda节点输出走分支路由
	err = graph.AddBranch("lambda", branch)
	if err != nil {
		panic(err)
	}

	// 3. 搭建边链路
	// 起点 -> lambda
	err = graph.AddEdge(compose.START, "lambda")
	if err != nil {
		panic(err)
	}
	// 两个分支节点分别流向llm
	err = graph.AddEdge("lambda1", "llm")
	if err != nil {
		panic(err)
	}
	err = graph.AddEdge("lambda2", "llm")
	if err != nil {
		panic(err)
	}
	// llm流向终点
	err = graph.AddEdge("llm", compose.END)
	if err != nil {
		panic(err)
	}

	compile, err := graph.Compile(ctx)
	if err != nil {
		panic(err)
	}
	output, err := compile.Stream(ctx, map[string]string{
		"role":    "aojiao",
		"content": "你这么傲娇，我不喜欢你了",
	})
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
	// 流式接收循环结束后，输出换行，
}










