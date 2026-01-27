package partygame

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 自定义错误类型
var (
	ErrSessionNotFound    = errors.New("游戏会话未找到")
	ErrPlayerNotFound     = errors.New("玩家未找到")
	ErrInvalidOperation   = errors.New("无效操作")
	ErrSessionExpired     = errors.New("游戏会话已过期")
	ErrInsufficientPlayers = errors.New("玩家数量不足")
	ErrPlayerInSession    = errors.New("玩家已在游戏中")
	ErrMaxPlayersReached   = errors.New("达到最大玩家数量")
	ErrNotYourTurn        = errors.New("尚未轮到你")
	ErrCartridgeEmpty     = errors.New("弹夹已空")
)

// 错误处理器接口
type ErrorHandler interface {
	HandleError(ctx *zero.Ctx, err error)
}

// 默认错误处理器
type DefaultErrorHandler struct{}

func (h *DefaultErrorHandler) HandleError(ctx *zero.Ctx, err error) {
	logrus.Errorf("[PartyGame]错误处理: %v", err)
	
	// 根据错误类型返回不同的用户消息
	switch {
	case errors.Is(err, ErrSessionNotFound):
		sendErrorMessage(ctx, "游戏会话未找到，请先创建游戏")
	case errors.Is(err, ErrPlayerNotFound):
		sendErrorMessage(ctx, "玩家信息未找到")
	case errors.Is(err, ErrInvalidOperation):
		sendErrorMessage(ctx, "操作无效，请检查输入")
	case errors.Is(err, ErrSessionExpired):
		sendErrorMessage(ctx, "游戏会话已过期，请重新开始")
	case errors.Is(err, ErrInsufficientPlayers):
		sendErrorMessage(ctx, "玩家数量不足，至少需要2名玩家")
	case errors.Is(err, ErrPlayerInSession):
		sendErrorMessage(ctx, "你已经在游戏中")
	case errors.Is(err, ErrMaxPlayersReached):
		sendErrorMessage(ctx, "达到最大玩家数量")
	case errors.Is(err, ErrNotYourTurn):
		sendErrorMessage(ctx, "尚未轮到你")
	case errors.Is(err, ErrCartridgeEmpty):
		sendErrorMessage(ctx, "弹夹已空")
	default:
		sendErrorMessage(ctx, "发生错误: "+err.Error())
	}
}

// 全局错误处理器实例
var defaultErrorHandler = &DefaultErrorHandler{}

// 包装函数调用，自动处理错误
func safeCall(ctx *zero.Ctx, fn func() error) {
	if err := fn(); err != nil {
		defaultErrorHandler.HandleError(ctx, err)
	}
}

// 包装函数调用，返回结果并自动处理错误
func safeCallWithResult[T any](ctx *zero.Ctx, fn func() (T, error)) T {
	result, err := fn()
	if err != nil {
		defaultErrorHandler.HandleError(ctx, err)
		var zeroValue T
		return zeroValue
	}
	return result
}

// 验证游戏会话
func validateSession(ctx *zero.Ctx, session Session) error {
	if session.GroupID == 0 {
		return ErrSessionNotFound
	}
	
	if session.IsExpired() {
		return ErrSessionExpired
	}
	
	if len(session.Users) < 2 {
		return ErrInsufficientPlayers
	}
	
	return nil
}

// 验证玩家操作权限
func validatePlayerAction(ctx *zero.Ctx, session Session, userID int64) error {
	if !session.IsPlayerInSession(userID) {
		return ErrPlayerNotFound
	}
	
	if !session.IsPlayerTurn(userID) {
		return ErrNotYourTurn
	}
	
	return nil
}

// 验证玩家加入权限
func validatePlayerJoin(session Session, userID int64) error {
	if session.IsPlayerInSession(userID) {
		return ErrPlayerInSession
	}
	
	if int(session.Max) <= session.GetUserCount() {
		return ErrMaxPlayersReached
	}
	
	return nil
}

// 获取错误详情（用于日志记录）
func getErrorDetails(err error, context string) string {
	return fmt.Sprintf("context: %s, error: %v", context, err)
}

// 错误恢复函数
func recoverWithError(ctx *zero.Ctx, context string) {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic recovered: %v", r)
		logrus.Errorf("[PartyGame]Panic in %s: %v", context, err)
		sendErrorMessage(ctx, "系统发生错误，请稍后重试")
	}
}

// 验证输入参数
func validateInput[T comparable](value T, validValues []T, paramName string) error {
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}
	return fmt.Errorf("%s 无效，有效值为: %v", paramName, validValues)
}

// 验证字符串输入
func validateStringInput(input string, minLength, maxLength int, fieldName string) error {
	if len(input) < minLength {
		return fmt.Errorf("%s 长度不能少于 %d 个字符", fieldName, minLength)
	}
	if len(input) > maxLength {
		return fmt.Errorf("%s 长度不能超过 %d 个字符", fieldName, maxLength)
	}
	return nil
}

// 验证数值范围
func validateNumberRange(value, min, max int64, fieldName string) error {
	if value < min {
		return fmt.Errorf("%s 不能小于 %d", fieldName, min)
	}
	if value > max {
		return fmt.Errorf("%s 不能大于 %d", fieldName, max)
	}
	return nil
}

// 错误消息国际化支持
type ErrorMessage struct {
	Chinese string
	English string
}

// 错误消息映射
var errorMessages = map[string]ErrorMessage{
	ErrSessionNotFound.Chinese:    {Chinese: "游戏会话未找到", English: "Game session not found"},
	ErrPlayerNotFound.Chinese:     {Chinese: "玩家未找到", English: "Player not found"},
	ErrInvalidOperation.Chinese:   {Chinese: "无效操作", English: "Invalid operation"},
	ErrSessionExpired.Chinese:     {Chinese: "游戏会话已过期", English: "Session expired"},
	ErrInsufficientPlayers.Chinese: {Chinese: "玩家数量不足", English: "Insufficient players"},
	ErrPlayerInSession.Chinese:    {Chinese: "玩家已在游戏中", English: "Player already in session"},
	ErrMaxPlayersReached.Chinese:  {Chinese: "达到最大玩家数量", English: "Max players reached"},
	ErrNotYourTurn.Chinese:        {Chinese: "尚未轮到你", English: "Not your turn"},
	ErrCartridgeEmpty.Chinese:     {Chinese: "弹夹已空", English: "Cartridge empty"},
}

// 获取本地化的错误消息
func getLocalizedErrorMessage(err error, language string) string {
	errStr := err.Error()
	if msg, exists := errorMessages[errStr]; exists {
		if strings.ToLower(language) == "en" || strings.ToLower(language) == "english" {
			return msg.English
		}
		return msg.Chinese
	}
	return err.Error()
}

// 批量错误处理
func handleBatchErrors(ctx *zero.Ctx, errors []error) {
	if len(errors) == 0 {
		return
	}
	
	errorMsgs := make([]string, len(errors))
	for i, err := range errors {
		errorMsgs[i] = err.Error()
	}
	
	logrus.Errorf("[PartyGame]Batch errors: %v", errorMsgs)
	sendErrorMessage(ctx, fmt.Sprintf("发生 %d 个错误，请检查操作", len(errors)))
}

// 带上下文的错误处理
type ContextError struct {
	Context string
	Err     error
}

func (ce *ContextError) Error() string {
	return fmt.Sprintf("%s: %v", ce.Context, ce.Err)
}

func (ce *ContextError) Unwrap() error {
	return ce.Err
}

// 创建带上下文的错误
func newContextError(context string, err error) *ContextError {
	return &ContextError{
		Context: context,
		Err:     err,
	}
}