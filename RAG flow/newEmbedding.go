package main

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ollama"
)

func EmbeddingClient(ctx context.Context) *ollama.Embedder {
	embedder, err := ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
		BaseURL: "http://localhost:11434",
		Model:   "bge-m3",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	return embedder
}
