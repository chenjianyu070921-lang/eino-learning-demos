package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// getEmbedding 获取文本向量
func getEmbedding(text string) ([]float64, error) {
	// 连接本地 ollama 服务
	u, err := url.Parse("http://127.0.0.1:11434")
	if err != nil {
		return nil, err
	}
	client := api.NewClient(u, &http.Client{})
	ctx := context.Background()

	req := &api.EmbeddingRequest{
		Model:  "bge-m3", // 你的嵌入模型
		Prompt: text,
	}

	resp, err := client.Embeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Embedding, nil
}

func main() {
	vector, err := getEmbedding("测试文本，用于RAG知识库检索")
	if err != nil {
		log.Fatalf("获取向量失败: %v", err)
	}

	fmt.Printf("向量维度: %d\n", len(vector))
	fmt.Printf("前5个向量值: %v\n", vector[:5])
}
