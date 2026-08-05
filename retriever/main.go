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
	// 创建检索器（Retriever）：给定一个问题，先在内部把问题向量化，再去 Milvus 里找最相似的文档。
	newRetriever, err := retriever.NewRetriever(ctx, &retriever.RetrieverConfig{
		Client:       client,
		Collection:   "test", // 要查询的集合名（必须先由 indexer 建过同名的集合，否则查不到）
		VectorField:  "vector", // 存向量的字段名（和 indexer 写入时一致）
		OutputFields: []string{"content", "metadata"}, // 命中后要返回的字段
		TopK:         2,                               // 只返回最相关的 2 条
		SearchMode:   search_mode.NewApproximate(retriever.COSINE), // 近似检索 + 余弦相似度
		Embedding:    embedder, // 用来把查询文本向量化（维度要和入库时一致）
	})
	if err != nil {
		panic(err)
	}
	// 检索：把"原神"变成向量，去集合里找最相近的 2 个文档块
	docs, err := newRetriever.Retrieve(ctx, "原神")
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		fmt.Println(doc)
	}
}
