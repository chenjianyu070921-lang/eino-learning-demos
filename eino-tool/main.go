package main

import (
	"context"
	"encoding/json"
	"fmt"
)

func main() {
	// 初始化工具：CreateTool() 返回一个“可被大模型调用”的 InvokableTool（定义在 tool.go）。
	invokableTool := CreateTool()
	ctx := context.Background()
	// Info：提取工具的“说明书”（名字 + 描述 + 参数 schema）。
	// 大模型就是靠这段元数据判断“什么时候该调这个工具、传什么参数”。
	info, err := invokableTool.Info(ctx)
	if err != nil {
		fmt.Printf("提取元数据失败: %v\n", err)
		return
	}
	fmt.Printf("工具名称: %s\n工具描述: %s\n", info.Name, info.Desc)
	// 构建调用参数：Eino 工具统一用 JSON 字符串传参。
	// 这里把结构体序列化成 JSON（复杂参数推荐，避免手写拼错）。
	params := InputParams{Name: "deepseek"}
	jsonBytes, _ := json.Marshal(params)
	jsonParams := string(jsonBytes)
	// 调用工具：InvokableRun 内部会把 JSON 参数反序列化、执行 GetGame 函数、返回结果字符串。
	run, err := invokableTool.InvokableRun(ctx, jsonParams)
	if err != nil {
		fmt.Printf("工具执行失败: %v\n", err)
		return
	}
	fmt.Printf("执行结果: %s\n", run)
}
