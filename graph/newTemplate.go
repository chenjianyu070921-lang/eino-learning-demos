package main

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func ChatTemplateClient(ctx context.Context) *prompt.DefaultChatTemplate {
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是资料问答助手，请严格依据下面提供的资料作答；资料中没有的内容，不要编造，直接说不知道。"),
		&schema.Message{
			Role:    schema.User,
			Content: "{task}",
		},
	)
	return template
}
