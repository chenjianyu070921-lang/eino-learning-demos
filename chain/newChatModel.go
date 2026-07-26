package main

import (
	"context"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ollama"
)

func ChatModelClient(ctx context.Context) *ollama.ChatModel {
	model, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: "http://127.0.0.1:11434", // ollama默认地址，不要加/v1
		Model:   "qwen3.5:9b",
		Timeout: 60 * time.Second, // HTTP请求超时 60秒
	})
	if err != nil {
		panic(err)
	}
	return model
}
