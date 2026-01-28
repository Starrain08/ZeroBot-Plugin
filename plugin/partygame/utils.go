package partygame

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 全局随机数生成器
var (
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// 从消息中解析at的用户ID
func parseTargetUserID(ctx *zero.Ctx) int64 {
	targetUserID := ctx.Event.UserID
	if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
		uid, err := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		if err == nil {
			targetUserID = uid
		}
	}
	return targetUserID
}

// 发送错误消息
func sendErrorMessage(ctx *zero.Ctx, msg string) {
	// 验证消息内容
	if err := validateStringInput(msg, 1, MaxMessageLength, "ErrorMessage"); err != nil {
		logrus.Errorf("[PartyGame]错误消息验证失败: %v", err)
		msg = "系统错误"
	}

	ctx.SendChain(message.Text(msg))
}

// 发送成功消息
func sendSuccessMessage(ctx *zero.Ctx, msg string) {
	// 验证消息内容
	if err := validateStringInput(msg, 1, MaxMessageLength, "SuccessMessage"); err != nil {
		logrus.Errorf("[PartyGame]成功消息验证失败: %v", err)
		msg = "操作成功"
	}

	ctx.SendChain(message.Text(msg))
}

// 发送带at的成功消息
func sendSuccessMessageWithAt(ctx *zero.Ctx, targetUserID int64, msg string) {
	if err := ValidateOnly(targetUserID); err != nil {
		logrus.Errorf("[PartyGame]目标用户ID验证失败: %v", err)
		targetUserID = ctx.Event.UserID
	}

	if err := validateStringInput(msg, 1, MaxMessageLength, "SuccessMessageWithAt"); err != nil {
		logrus.Errorf("[PartyGame]带at的成功消息验证失败: %v", err)
		msg = "操作成功"
	}

	ctx.SendChain(message.At(targetUserID), message.Text(msg))
}

// 检查是否为管理员权限
func isAdmin(ctx *zero.Ctx) bool {
	return zero.SuperUserPermission(ctx) || ctx.GetGroupMemberInfo(ctx.Event.GroupID, ctx.Event.UserID, true).Get("role").String() != "member"
}

// 随机选择数组中的元素
func randomChoice[T any](items []T) T {
	if len(items) == 0 {
		var zeroValue T
		return zeroValue
	}

	if err := ValidateOnly(items); err != nil {
		logrus.Warnf("[PartyGame]随机选择输入验证失败: %v", err)
		var zeroValue T
		return zeroValue
	}

	return items[randSource.Intn(len(items))]
}

// RandomChoice 导出函数，用于测试
func RandomChoice[T any](items []T) T {
	return randomChoice(items)
}

// Unique 导出函数，用于测试
func Unique[T comparable](items []T) []T {
	return unique(items)
}

// GenerateRouletteCartridges 导出函数，用于测试
func GenerateRouletteCartridges() []int {
	return generateRouletteCartridges()
}

// 切片去重
func unique[T comparable](items []T) []T {
	if len(items) == 0 {
		return items
	}

	seen := make(map[T]bool)
	result := make([]T, 0, len(items))

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// 安全的字符串拼接
func safeStringConcat(parts ...string) string {
	var result strings.Builder

	for _, part := range parts {
		result.WriteString(part)
	}

	return result.String()
}

// 带超时的操作执行
func withTimeout[T any](fn func() T, timeout time.Duration) (T, error) {
	done := make(chan T, 1)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic: %v", r)
			}
		}()
		result := fn()
		done <- result
	}()

	select {
	case result := <-done:
		return result, nil
	case err := <-errChan:
		var zeroValue T
		return zeroValue, err
	case <-time.After(timeout):
		var zeroValue T
		return zeroValue, fmt.Errorf("operation timed out after %v", timeout)
	}
}

// 线程安全的计数器
type SafeCounter struct {
	count int
	mu    sync.Mutex
}

func (sc *SafeCounter) Increment() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.count++
}

func (sc *SafeCounter) Decrement() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.count--
}

func (sc *SafeCounter) Get() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.count
}

func (sc *SafeCounter) Reset() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.count = 0
}
