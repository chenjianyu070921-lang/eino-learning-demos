package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// getEmbedding 获取文本向量（直接用 Ollama 官方 Go SDK，不经过 Eino 封装）
// 返回 []float64，即这段文本的数字向量表示。
func getEmbedding(text string) ([]float64, error) {
	// 连接本地 ollama 服务（前提：本机装了 Ollama 且拉取了 bge-m3）
	u, err := url.Parse("http://127.0.0.1:11434")
	if err != nil {
		return nil, err
	}
	client := api.NewClient(u, &http.Client{})
	ctx := context.Background()

	// 构造 embedding 请求：指定模型名和要向量化的文本
	req := &api.EmbeddingRequest{
		Model:  "bge-m3", // 你的嵌入模型（1024 维）
		Prompt: text,
	}

	// 调用 Embeddings 接口，返回向量
	resp, err := client.Embeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Embedding, nil
}

func main() {
	// 对一句测试文本做向量化，并打印维度和前几个数值，验证 embedding 可用
	vector, err := getEmbedding("测试文本，用于RAG知识库检索")
	if err != nil {
		log.Fatalf("获取向量失败: %v", err)
	}

	fmt.Printf("向量维度: %d\n", len(vector))
	fmt.Printf("前5个向量值: %v\n", vector[:5])
}
