package main

import (
	"context"
	"encoding/json"
	"fmt"
)

func main() {
	//初始化工具
	invokableTool := CreateTool()
	ctx := context.Background()
	//提取元数据
	info, err := invokableTool.Info(ctx)
	if err != nil {
		fmt.Printf("提取元数据失败: %v\n", err)
		return
	}
	fmt.Printf("工具名称: %s\n工具描述: %s\n", info.Name, info.Desc)
	//构建消息
	//结构体转 JSON（复杂参数推荐，避免手写拼错）
	params := InputParams{Name: "deepseek"}
	jsonBytes, _ := json.Marshal(params)
	jsonParams := string(jsonBytes)
	//调用工具

	run, err := invokableTool.InvokableRun(ctx, jsonParams)
	if err != nil {
		fmt.Printf("工具执行失败: %v\n", err)
		return
	}
	fmt.Printf("执行结果: %s\n", run)
}
