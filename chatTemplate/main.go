package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("加载.env文件失败，请确认文件在项目根目录：" + err.Error())
	}
	//初始化上下文
	ctx := context.Background()
	//初始化大模型
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		BaseURL:  os.Getenv("ARK_BASE_URL"),
		APIKey:   os.Getenv("ARK_API_KEY"),
		Model:    os.Getenv("ARK_MODEL_ID"),
		Thinking: &arkModel.Thinking{Type: arkModel.ThinkingTypeDisabled},
	})
	if err != nil {
		panic(err)
	}
	// 建立 prompt 模板。
	// FromMessages 的第一个参数 schema.FString 表示：用 {变量名} 占位符格式化。
	// System / User 消息里的 {role}、{task} 都是占位符，运行时会用 params 里的真实值替换。
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
	// params 里 key 必须和模板占位符名字一致：role → {role}，task → {task}
	params := map[string]any{
		"role": "JianYu先生",
		"task": "eino九大组件哪个最重要",
	}
	// Format：把占位符替换成真实值，返回填好的 []*schema.Message
	result, err := template.Format(ctx, params)
	if err != nil {
		panic(err)
	}
	msg, err := model.Generate(ctx, result)
	if err != nil {
		panic(err)
	}
	answer := msg.Content
	fmt.Println("模型回答：\n", answer)

}
