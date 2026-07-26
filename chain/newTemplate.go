package main

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func ChatTemplateClient(ctx context.Context) *prompt.DefaultChatTemplate {
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是{role}，专业Go与Eino框架技术讲师。\n"+
			"硬性要求：\n"+
			"1. 禁止任何角色扮演、剧情、比喻、悬疑台词，不要加括号动作描述；\n"+
			"2. 直接清晰、分点讲解技术内容，客观直白；\n"+
			"3. 用户询问Eino框架时，完整讲解基础使用流程、核心组件、运行步骤；\n"+
			"4. 不回避技术问题，不编造隐喻故事。"),
		&schema.Message{
			Role:    schema.User,
			Content: "请帮帮我，JianYu先生，{task}",
		},
	)
	return template
}
