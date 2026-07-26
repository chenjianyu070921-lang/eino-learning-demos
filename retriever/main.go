package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus2/search_mode"
	"github.com/joho/godotenv"

	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus2"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "127.0.0.1:19530",
		DBName:  "default",
	})
	if err != nil {
		panic(err)
	}
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
	newRetriever, err := retriever.NewRetriever(ctx, &retriever.RetrieverConfig{
		Client:       client,
		Collection:   "test",
		VectorField:  "vector",
		OutputFields: []string{"content", "metadata"},
		TopK:         2,
		SearchMode:   search_mode.NewApproximate(retriever.COSINE),
		Embedding:    embedder,
	})
	if err != nil {
		panic(err)
	}
	docs, err := newRetriever.Retrieve(ctx, "原神")
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		fmt.Println(doc)
	}
}
