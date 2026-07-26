package main

import (
	"context"
	"fmt"
	"strings"

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
	name := strings.ToLower(strings.TrimSpace(params.Name))
	fmt.Printf("[工具执行] get_llm_url，参数 name=%q\n", name)

	gameSet := []Game{
		{Name: "chatgpt", Url: "https://chatgpt.com/"},
		{Name: "doubao", Url: "https://www.doubao.com/chat/"},
		{Name: "deepseek", Url: "https://www.deepseek.com/"},
		{Name: "原神", Url: "https://ys.mihoyo.com/main/?af_adset=1"},
	}
	for _, game := range gameSet {
		if game.Name == name {
			fmt.Printf("[工具结果] %s\n", game.Url)
			return game.Url, nil
		}
	}
	return "", fmt.Errorf("未找到名称为 %q 的官方网址", params.Name)
}
func CreateTool() tool.InvokableTool {
	getGameTool := utils.NewTool(&schema.ToolInfo{
		Name:  "get_llm_url",                //给工具起一个唯一性的名称
		Desc:  "根据名称查询官方访问链接。支持 chatgpt、doubao、deepseek 和原神；当用户询问这些网站的官方网址时应调用此工具。",
		Extra: nil,                          //扩展信息
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"name": {
					Type:     schema.String,
					Desc:     "网站或产品名称，可选值：chatgpt、doubao、deepseek、原神",
					Required: true,
				},
			},
		),
	}, GetGame)
	return getGameTool
}
