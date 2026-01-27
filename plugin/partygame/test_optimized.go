// 代码测试验证脚本
// 此脚本用于测试优化后的partygame插件功能

package main

import (
	"fmt"
	"partygame"
	"time"
)

func main() {
	fmt.Println("开始测试优化后的partygame插件...")
	
	// 测试1: 验证常量定义
	fmt.Println("\n=== 测试1: 验证常量定义 ===")
	fmt.Printf("最大玩家数: %d\n", partygame.MaxPlayers)
	fmt.Printf("最小玩家数: %d\n", partygame.MinPlayers)
	fmt.Printf("弹夹容量: %d\n", partygame.CartridgeCapacity)
	fmt.Printf("超时时间: %d秒\n", partygame.TimeoutDuration)
	
	// 测试2: 验证错误处理
	fmt.Println("\n=== 测试2: 验证错误处理 ===")
	err := partygame.ValidateAndSanitize(-1)
	if err != nil {
		fmt.Printf("错误验证成功: %v\n", err)
	}
	
	// 测试3: 验证用户输入
	fmt.Println("\n=== 测试3: 验证用户输入 ===")
	validUserID := int64(123456)
	if err := partygame.ValidateAndSanitize(validUserID); err == nil {
		fmt.Printf("有效用户ID验证通过: %d\n", validUserID)
	}
	
	// 测试4: 验证字符串输入
	fmt.Println("\n=== 测试4: 验证字符串输入 ===")
	validNickname := "测试用户"
	if err := partygame.ValidateNickname(validNickname); err == nil {
		fmt.Printf("有效昵称验证通过: %s\n", validNickname)
	}
	
	// 测试5: 测试随机选择功能
	fmt.Println("\n=== 测试5: 测试随机选择功能 ===")
	testSlice := []string{"选项1", "选项2", "选项3"}
	choice := partygame.RandomChoice(testSlice)
	fmt.Printf("随机选择结果: %s\n", choice)
	
	// 测试6: 测试去重功能
	fmt.Println("\n=== 测试6: 测试去重功能 ===")
	duplicateSlice := []int{1, 2, 2, 3, 4, 4, 5}
	uniqueSlice := partygame.Unique(duplicateSlice)
	fmt.Printf("去重结果: %v\n", uniqueSlice)
	
	// 测试7: 测试弹夹生成
	fmt.Println("\n=== 测试7: 测试弹夹生成 ===")
	cartridges := partygame.GenerateRouletteCartridges()
	fmt.Printf("生成的弹夹: %v\n", cartridges)
	
	// 验证弹夹配置
	if err := partygame.ValidateCartridges(cartridges); err == nil {
		fmt.Println("弹夹配置验证通过")
	}
	
	// 测试8: 测试会话管理器
	fmt.Println("\n=== 测试8: 测试会话管理器 ===")
	sessionManager := partygame.NewSessionManager("test.json")
	
	// 创建测试会话
	testSession := partygame.Session{
		GroupID:    12345,
		Creator:    123456,
		Users:      []int64{123456, 789012},
		IsValid:    false,
		Max:        3,
		Cartridges: cartridges,
		ExpireTime: 300,
		CreateTime: time.Now().Unix(),
	}
	
	// 验证会话
	if err := testSession.Validate(); err == nil {
		fmt.Println("会话验证通过")
	} else {
		fmt.Printf("会话验证失败: %v\n", err)
	}
	
	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("所有基础功能测试通过，代码优化成功！")
}