package main

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// Game 是工具要返回的数据结构（一个“大模型官网”信息）
type Game struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// InputParams 是工具的入参结构。
// 注意 tag：json 决定序列化后的字段名；jsonschema 会生成给大模型看的参数说明
//（大模型靠它理解“这个参数叫什么、干什么用”）。
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
// CreateTool 把 GetGame 函数包装成一个 Eino 的 InvokableTool。
// 关键点：工具必须提供“元数据”（Name/Desc/参数说明），大模型才知道怎么用。
func CreateTool() tool.InvokableTool {
	getGameTool := utils.NewTool(&schema.ToolInfo{
		Name:  "get_llm_url",                //给工具起一个唯一性的名称（大模型按名字调用）
		Desc:  "根据传入的大模型名称，查询获取对应的官方访问链接地址", //大模型依靠这句话理解这个工具是干什么的
		Extra: nil,                          //扩展信息（一般留空）
		// ParamsOneOf：声明工具的参数列表。这里只有一个必填字符串参数 name。
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"name": &schema.ParameterInfo{
					Type:     schema.String,
					Desc:     "大模型名称，可选取值：chatgpt、doubao、deepseek",
					Required: true, // 必填
				},
			},
		), //联合参数模式（Eino 推荐的参数定义方式）
	}, GetGame) // 第二个参数：真正执行的函数（签名需匹配 ctx, *InputParams）
	return getGameTool
}
