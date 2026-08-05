package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("程序开始")

	err := godotenv.Load()
	if err != nil {
		panic("加载.env文件失败：" + err.Error())
	}
	ctx := context.Background()
	// 创建 Embedder（把文本变成向量）。
	// 注意这里读的是 EMBEDDER 环境变量（即你的 embedding 模型 id，例如 ep-xxxx），
	// 和对话用的 ARK_MODEL_ID 是两回事。
	// APIType 设为 MultiModal：开启多模态模式（支持文本+图片等）。
	apiType := ark.APITypeMultiModal
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  os.Getenv("ARK_API_KEY"),
		Model:   os.Getenv("EMBEDDER"),
		APIType: &apiType, // 【重中之重】开启多模态模式！
	})
	if err != nil {
		log.Fatalf("NewEmbedder of ark error: %v", err)
		return
	}

	// EmbedStrings：把一组文本批量转成向量。每条文本对应一个 []float64 向量。
	// 向量的维度（长度）取决于模型，常见 1024 / 2048 等。
	vectors, err := embedder.EmbedStrings(ctx, []string{"hello", "how are you"})
	if err != nil {
		log.Fatalf("EmbedStrings of Ark failed, err=%v", err)
	}

	log.Printf("vectors : %v", vectors)
	// 打印向量的“维度”（每个向量有多少个数字）。RAG 入库和检索时维度必须一致。
	fmt.Println("embedding维度:", len(vectors[0]))
}
