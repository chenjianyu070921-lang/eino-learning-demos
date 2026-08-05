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

// State 是“图状态”：整张图运行期间一直存活的共享对象，任何节点都能读写。
// 这里只用 History 存两段“角色描述”，供不同分支读取。
// 注意：图的输入/输出类型都是 map[string]any（用 any 是为了在节点间灵活传值）。
type State struct {
	History map[string]any
}

// NewState 是 State 的初始化函数，通过 WithGenLocalState 注册给图。
// 每次运行 Compile 后的 Runnable，都会新建一份独立的 State。
func NewState(ctx context.Context) *State {
	return &State{History: make(map[string]any)}
}
func main() {
	if err := godotenv.Load(); err != nil {
		panic("加载.env文件失败：" + err.Error())
	}
	fmt.Println(".env加载成功")
	ctx := context.Background()

	// 创建带 State 的图：第二个泛型是输出类型 *schema.Message（最终给大模型）。
	// WithGenLocalState(NewState) 告诉框架“这张图有一个共享 State，用 NewState 初始化”。
	graph := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(NewState),
	)

	// lambda（路由节点）：根据输入 role，原样输出 {role, content} 供后面 branch 判断；
	// 同时用 ProcessState 往 State.History 写入两段角色描述（aojiao/keai 各一句）。
	// ProcessState：在节点内部安全地读写 State（框架会处理并发）。
	lamdba := compose.InvokableLambda(func(ctx context.Context,
		input map[string]any) (output map[string]any, err error) {
		//在节点内部处理state（写）
		_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
			state.History["aojiao_action"] = "我喜欢你"
			state.History["keai_action"] = "摸摸头"
			return nil
		})
		if input["role"] == "aojiao" {
			return map[string]any{"role": "aojiao", "content": input["content"]}, nil
		}
		if input["role"] == "keai" {
			return map[string]any{"role": "keai", "content": input["content"]}, nil
		}
		return map[string]any{"role": "user", "content": input["content"]}, nil
	})
	// lambda1（傲娇分支）：从 State 读出 aojiao_action 拼到用户问题后面，再构造 prompt 消息。
	// 这里演示“State 被另一个节点读取”——lambda 写入，这里读取，无需用边传递。
	aojiaoLamdba := compose.InvokableLambda(func(ctx context.Context,
		input map[string]any) (output []*schema.Message, err error) {
		//在节点内部处理state（读）
		_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
			content, _ := input["content"].(string)
			content += state.History["aojiao_action"].(string)
			input["content"] = content
			return nil
		})
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个高冷傲娇的大小姐，每次都会用傲娇的语气回答我的问题（你的内心是一个病娇）",
			},
			{
				Role:    schema.User,
				Content: input["content"].(string),
			},
		}, nil
	})

	// lambda2（可爱分支）：和傲娇分支逻辑一模一样，只是从 State 读 keai_action。
	keailambda := compose.InvokableLambda(func(ctx context.Context,
		input map[string]any) (output []*schema.Message, err error) {
		_ = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
			content, _ := input["content"].(string)
			content += state.History["keai_action"].(string)
			input["content"] = content
			return nil
		})
		return []*schema.Message{
			{
				Role:    schema.System,
				Content: "你是一个可爱的小女孩，每次都会用可爱的语气回答我的问题",
			},
			{
				Role:    schema.User,
				Content: input["content"].(string),
			},
		}, nil
	})
	// 建立分支（Branch）：根据输入 in["role"] 的值，把数据路由到 lambda1（傲娇）或 lambda2（可爱）。
	// 注意 in["role"] 是 any，先用 .(string) 断言成字符串再判断。
	branch := compose.NewGraphBranch(
		func(ctx context.Context, in map[string]any) (string, error) {
			role, ok := in["role"].(string)
			if !ok {
				return "", fmt.Errorf("role 类型转换失败")
			}
			switch role {
			case "aojiao":
				return "lambda1", nil
			case "keai":
				return "lambda2", nil
			default:
				return "lambda1", nil
			}
		},
		// 第二个参数：所有分支可跳转节点列表（框架据此校验分支输出必须落在这几个节点里）
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
	// 2. 绑定分支：lambda节点输出走分支路由。
	// ⚠️ 重要顺序：AddBranch 必须在 lambda1/lambda2 这两个目标节点“已加入图”之后调用，
	// 否则编译会报 "branch end node needs to be added to graph first"。
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
	output, err := compile.Stream(ctx, map[string]any{
		"role":    "keai",
		"content": "你好呀",
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
