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

	vectors, err := embedder.EmbedStrings(ctx, []string{"hello", "how are you"})
	if err != nil {
		log.Fatalf("EmbedStrings of Ark failed, err=%v", err)
	}

	log.Printf("vectors : %v", vectors)
	fmt.Println("embedding维度:", len(vectors[0]))
}
