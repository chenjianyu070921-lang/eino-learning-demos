package main

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type Game struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
type InputParams struct {
	Name string `json:"name" jsonschema:"description=the name of game"`
}

func GetGame(ctx context.Context, params *InputParams) (string, error) {
	GameSet := []Game{
		{Name: "chatgpt", Url: "https://chatgpt.com/"},
		{Name: "doubao", Url: "https://www.doubao.com/chat/"},
		{Name: "deepseek", Url: "https://www.deepseek.com/"},
	}
	for _, game := range GameSet {
		if game.Name == params.Name {
			return game.Url, nil
		}
	}
	return "", nil
}
func CreateTool() tool.InvokableTool {
	getGameTool := utils.NewTool(&schema.ToolInfo{
		Name:  "get_llm_url",                //给工具起一个唯一性的名称
		Desc:  "根据传入的大模型名称，查询获取对应的官方访问链接地址", //大模型依靠这句话理解这个工具是干什么的
		Extra: nil,                          //扩展信息
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"name": &schema.ParameterInfo{
					Type:     schema.String,
					Desc:     "大模型名称，可选取值：chatgpt、doubao、deepseek",
					Required: true,
				},
			},
		), //联合参数模式
	}, GetGame)
	return getGameTool
}
