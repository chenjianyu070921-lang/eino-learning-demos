package main

import (
	"context"

	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus2"
	"github.com/cloudwego/eino-ext/components/retriever/milvus2/search_mode"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// RetrieverClient 创建 milvus2 检索器。embedder 使用 embedding.Embedder 接口。
func RetrieverClient(ctx context.Context, newClient *milvusclient.Client, embedder embedding.Embedder) *retriever.Retriever {
	newRetriever, err := retriever.NewRetriever(ctx, &retriever.RetrieverConfig{
		Client:       newClient,
		Collection:   "personal_knowledge_base",
		VectorField:  "vector",
		OutputFields: []string{"content", "metadata"},
		TopK:         3,
		SearchMode:   search_mode.NewApproximate(retriever.COSINE),
		Embedding:    embedder,
	})
	if err != nil {
		panic(err)
	}
	return newRetriever
}
