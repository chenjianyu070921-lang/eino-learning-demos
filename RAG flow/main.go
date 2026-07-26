package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()
	milvusClient := MilvusClient(ctx)
	embedder := EmbeddingClient(ctx)
	chatModel := ChatModelClient(ctx)
	retriever := RetrieverClient(ctx, milvusClient, embedder)

	// 你的提问（实际使用时可改为从命令行/前端获取）
	query := "今天天气怎么样"

	// 1) 意图路由：判断这个问题是否需要查向量库
	needRAG, err := needRetrieve(ctx, chatModel, query)
	if err != nil {
		panic(err)
	}

	var systemMsg, userMsg string
	if needRAG {
		// 2) 需要查资料：检索 + 拼 context 进 user 消息
		docs, err := retriever.Retrieve(ctx, query)
		if err != nil {
			panic(err)
		}
		var sb strings.Builder
		for i, d := range docs {
			sb.WriteString(fmt.Sprintf("【资料%d】\n%s\n\n", i+1, d.Content))
		}
		systemMsg = "你是资料问答助手，请严格依据下面提供的资料作答；资料中没有的内容，不要编造，直接说不知道。"
		userMsg = fmt.Sprintf("资料：\n%s\n\n用户问题：%s", sb.String(), query)
	} else {
		// 3) 不需要查资料：直接把原问题丢给模型，完全不访问向量库
		systemMsg = "你是一个乐于助人的通用助手。"
		userMsg = query
	}

	messages := []*schema.Message{
		schema.SystemMessage(systemMsg),
		schema.UserMessage(userMsg),
	}

	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	for {
		recv, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}
		if recv != nil {
			_, _ = fmt.Print(recv.Content)
		}
	}
}

// needRetrieve 用模型做意图判断：这个问题是否需要查知识库资料。
// 返回 true => 走 RAG 检索；返回 false => 直接对话，不碰 Milvus。
func needRetrieve(ctx context.Context, cm *ark.ChatModel, query string) (bool, error) {
	judge := schema.SystemMessage("你是一个路由判断器。如果用户的问题需要参考知识库资料才能准确回答，只回复 YES；如果是闲聊、通用常识、或无需查资料就能回答的问题，只回复 NO。只输出 YES 或 NO，不要多余内容。")
	resp, err := cm.Generate(ctx, []*schema.Message{judge, schema.UserMessage(query)})
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.ToUpper(strings.TrimSpace(resp.Content)), "YES"), nil
}

//// 索引入库逻辑（需要时启用）：把 md 文件切分后写入 Milvus
//splitter, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
//	Headers: map[string]string{
//		"#":   "h1",
//		"##":  "h2",
//		"###": "h3",
//	},
//	TrimHeaders: true,
//})
//if err != nil {
//	panic(err)
//}
//content, err := os.ReadFile("互联网热梗大赏.md")
//if err != nil {
//	panic(err)
//}
//datas := []*schema.Document{
//	{
//		ID:      "1",
//		Content: string(content),
//	},
//}
//transform, err := splitter.Transform(ctx, datas)
//if err != nil {
//	panic(err)
//}
//for i, data := range transform {
//	data.ID = data.ID + strconv.Itoa(i)
//}
//ids, err := indexer.Store(ctx, transform)
//if err != nil {
//	panic(err)
//}
//fmt.Println(ids)
