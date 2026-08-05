package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
)

// Chain = 把组件串成一条直线（糖葫芦），数据从一头进、另一头出
func main() {
	err := godotenv.Load()
	if err != nil {
		panic("加载.env文件失败：" + err.Error())
	}

	fmt.Println(".env加载成功")
	ctx := context.Background()
	chatModelClient := ChatModelClient(ctx)
	templateClient := ChatTemplateClient(ctx)
	// 要点（新手必记）：
	// 1. 前一个节点的“输出类型”必须等于下一个节点的“输入类型”，否则编译报 mismatch。
	//    这里 ChatTemplate 输出 []*schema.Message，正好就是 ChatModel 的输入类型。
	// 2. Chain 是“直线糖葫芦”，不用像 Graph 那样手写 AddEdge / 连到 END。
	// 3. Chain 也能分支/并行，但本例是简单的 模板→模型 两步。
	// 4. Chain 可被嵌套复用（当成一个节点放进更大的 Graph 里）。
	chain := compose.NewChain[map[string]any, *schema.Message]().
		AppendChatTemplate(templateClient).
		AppendChatModel(chatModelClient)
	r, _ := chain.Compile(ctx)
	msg, _ := r.Invoke(ctx, map[string]any{
		"role": "JianYu先生",
		"task": "eino九大组件哪个最重要",
	})
	fmt.Println(msg.Content)
}

//第1步  创建：   chain := compose.NewChain[I, O]()
//第2步  加节点： chain.AppendChatTemplate(...).AppendChatModel(...)   // 链式拼接
//第3步  编译：   r, _ := chain.Compile(ctx)        // 变成可运行的 Runnable
//第4步  运行：   r.Invoke(ctx, input)              // 一进一出
//r.Stream(ctx, input)              // 一进多出（流式）
//r.Collect(ctx, reader)            // 多进一出
//r.Transform(ctx, reader)          // 多进多出（流进流出）
