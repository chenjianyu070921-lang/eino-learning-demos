package main

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

func ChatModelClient(ctx context.Context) *ark.ChatModel {
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		BaseURL:  os.Getenv("ARK_BASE_URL"),
		APIKey:   os.Getenv("ARK_API_KEY"),
		Model:    os.Getenv("ARK_MODEL_ID"),
		Thinking: &arkModel.Thinking{Type: arkModel.ThinkingTypeDisabled},
	})
	if err != nil {
		panic(err)
	}
	return model
}
