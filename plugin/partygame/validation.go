package partygame

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %v)", ve.Field, ve.Message, ve.Value)
}

// Validator 验证器接口
type Validator interface {
	Validate() error
}

// SessionValidator 会话验证器
type SessionValidator struct {
	Session Session
}

func (sv *SessionValidator) Validate() error {
	if sv.Session.GroupID <= 0 {
		return &ValidationError{
			Field:   "GroupID",
			Message: "group ID must be positive",
			Value:   sv.Session.GroupID,
		}
	}

	if sv.Session.Creator <= 0 {
		return &ValidationError{
			Field:   "Creator",
			Message: "creator ID must be positive",
			Value:   sv.Session.Creator,
		}
	}

	if sv.Session.Max < MinPlayers || sv.Session.Max > MaxPlayers {
		return &ValidationError{
			Field:   "Max",
			Message: fmt.Sprintf("max players must be between %d and %d", MinPlayers, MaxPlayers),
			Value:   sv.Session.Max,
		}
	}

	if len(sv.Session.Users) > int(sv.Session.Max) {
		return &ValidationError{
			Field:   "Users",
			Message: fmt.Sprintf("number of users (%d) exceeds max players (%d)", len(sv.Session.Users), sv.Session.Max),
			Value:   len(sv.Session.Users),
		}
	}

	for i, userID := range sv.Session.Users {
		if userID <= 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("Users[%d]", i),
				Message: "user ID must be positive",
				Value:   userID,
			}
		}
		// 检查重复用户
		for j := i + 1; j < len(sv.Session.Users); j++ {
			if userID == sv.Session.Users[j] {
				return &ValidationError{
					Field:   fmt.Sprintf("Users[%d]", i),
					Message: "duplicate user ID",
					Value:   userID,
				}
			}
		}
	}

	return nil
}

// UserValidator 用户验证器
type UserValidator struct {
	UserID   int64
	Nickname string
	GroupID  int64
}

func (uv *UserValidator) Validate() error {
	if uv.UserID <= 0 {
		return &ValidationError{
			Field:   "UserID",
			Message: "user ID must be positive",
			Value:   uv.UserID,
		}
	}

	if uv.GroupID <= 0 {
		return &ValidationError{
			Field:   "GroupID",
			Message: "group ID must be positive",
			Value:   uv.GroupID,
		}
	}

	if err := validateStringInput(uv.Nickname, MinNicknameLength, MaxNicknameLength, "Nickname"); err != nil {
		return err
	}

	return nil
}

// MessageValidator 消息验证器
type MessageValidator struct {
	Message string
	Type    string
	Length  int
}

func (mv *MessageValidator) Validate() error {
	if mv.Message == "" {
		return &ValidationError{
			Field:   "Message",
			Message: "message cannot be empty",
			Value:   mv.Message,
		}
	}

	if err := validateStringInput(mv.Message, 1, MaxMessageLength, "Message"); err != nil {
		return err
	}

	validTypes := []string{MsgTypeText, MsgTypeAt, MsgTypeImage, MsgTypeFace}
	if err := validateInput(mv.Type, validTypes, "Message Type"); err != nil {
		return err
	}

	if mv.Length < 0 || mv.Length > MaxMessageLength {
		return &ValidationError{
			Field:   "Length",
			Message: fmt.Sprintf("message length must be between 0 and %d", MaxMessageLength),
			Value:   mv.Length,
		}
	}

	return nil
}

// GameConfigValidator 游戏配置验证器
type GameConfigValidator struct {
	MaxPlayers int64
	Timeout    time.Duration
	MaxRetries int
	EnableLog  bool
}

func (gcv *GameConfigValidator) Validate() error {
	if gcv.MaxPlayers < MinPlayers || gcv.MaxPlayers > MaxPlayers {
		return &ValidationError{
			Field:   "MaxPlayers",
			Message: fmt.Sprintf("max players must be between %d and %d", MinPlayers, MaxPlayers),
			Value:   gcv.MaxPlayers,
		}
	}

	if gcv.Timeout <= 0 {
		return &ValidationError{
			Field:   "Timeout",
			Message: "timeout must be positive",
			Value:   gcv.Timeout,
		}
	}

	if gcv.MaxRetries < 0 || gcv.MaxRetries > MaxRetryAttempts {
		return &ValidationError{
			Field:   "MaxRetries",
			Message: fmt.Sprintf("max retries must be between 0 and %d", MaxRetryAttempts),
			Value:   gcv.MaxRetries,
		}
	}

	return nil
}

// 全局验证函数
// ValidateQQNumber 验证QQ号码
func ValidateQQNumber(qq string) error {
	matched, _ := regexp.MatchString(RegexQQNumber, qq)
	if !matched {
		return &ValidationError{
			Field:   "QQ",
			Message: "invalid QQ number format",
			Value:   qq,
		}
	}
	return nil
}

// ValidateNickname 验证昵称
func ValidateNickname(nickname string) error {
	return validateStringInput(nickname, MinNicknameLength, MaxNicknameLength, "Nickname")
}

// ValidateAction 验证游戏动作
func ValidateAction(action string) error {
	if action == "" {
		return &ValidationError{
			Field:   "Action",
			Message: "action cannot be empty",
			Value:   action,
		}
	}

	validActions := []string{"create", "join", "start", "fire", "terminate", "truth", "dare"}
	if err := validateInput(action, validActions, "Action"); err != nil {
		return err
	}

	return nil
}

// ValidateGameChoice 验证游戏选择
func ValidateGameChoice(choice string) error {
	if choice == "" {
		return &ValidationError{
			Field:   "Choice",
			Message: "choice cannot be empty",
			Value:   choice,
		}
	}

	if err := validateStringInput(choice, 2, 10, "Choice"); err != nil {
		return err
	}

	validChoices := []string{"truth", "dare", "真心话", "大冒险"}
	if err := validateInput(choice, validChoices, "Choice"); err != nil {
		return err
	}

	return nil
}

// ValidateTimeout 验证超时时间
func ValidateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return &ValidationError{
			Field:   "Timeout",
			Message: "timeout must be positive",
			Value:   timeout,
		}
	}

	if timeout > time.Hour*24 {
		return &ValidationError{
			Field:   "Timeout",
			Message: "timeout cannot exceed 24 hours",
			Value:   timeout,
		}
	}

	return nil
}

// ValidateFilepath 验证文件路径
func ValidateFilepath(filepath string) error {
	if filepath == "" {
		return &ValidationError{
			Field:   "Filepath",
			Message: "filepath cannot be empty",
			Value:   filepath,
		}
	}

	// 检查路径长度
	if len(filepath) > 260 {
		return &ValidationError{
			Field:   "Filepath",
			Message: "filepath too long (max 260 characters)",
			Value:   filepath,
		}
	}

	// 检查危险字符
	dangerousChars := []string{"..", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(filepath, char) {
			return &ValidationError{
				Field:   "Filepath",
				Message: fmt.Sprintf("filepath contains dangerous character: %s", char),
				Value:   filepath,
			}
		}
	}

	return nil
}

// ValidateJSON 验证JSON数据
func ValidateJSON(data []byte, schema interface{}) error {
	if len(data) == 0 {
		return &ValidationError{
			Field:   "JSON",
			Message: "JSON data cannot be empty",
			Value:   data,
		}
	}

	if len(data) > MaxFileSize {
		return &ValidationError{
			Field:   "JSON",
			Message: fmt.Sprintf("JSON data too large (max %d bytes)", MaxFileSize),
			Value:   len(data),
		}
	}

	// 这里可以添加更复杂的JSON schema验证
	return nil
}

// ValidatePermission 验证权限
func ValidatePermission(role string, requiredRole string) error {
	validRoles := []string{RoleCreator, RolePlayer, RoleAdmin, RoleSpectator}

	if err := validateInput(role, validRoles, "Role"); err != nil {
		return err
	}

	roleHierarchy := map[string]int{
		RoleSpectator: 0,
		RolePlayer:    1,
		RoleCreator:   2,
		RoleAdmin:     3,
	}

	userLevel := roleHierarchy[role]
	requiredLevel := roleHierarchy[requiredRole]

	if userLevel < requiredLevel {
		return &ValidationError{
			Field:   "Permission",
			Message: fmt.Sprintf("permission denied, required role: %s", requiredRole),
			Value:   role,
		}
	}

	return nil
}

// ValidateGameState 验证游戏状态
func ValidateGameState(state string) error {
	validStates := []string{GameStateCreated, GameStateActive, GameStateFinished, GameStateExpired}

	return validateInput(state, validStates, "GameState")
}

// ValidateCartridges 验证弹夹配置
func ValidateCartridges(cartridges []int) error {
	if len(cartridges) != CartridgeCapacity {
		return &ValidationError{
			Field:   "Cartridges",
			Message: fmt.Sprintf("cartridges length must be %d", CartridgeCapacity),
			Value:   len(cartridges),
		}
	}

	bulletCount := 0
	for _, pos := range cartridges {
		if pos != 0 && pos != 1 {
			return &ValidationError{
				Field:   fmt.Sprintf("Cartridges[%d]", pos),
				Message: "cartridge position must be 0 or 1",
				Value:   pos,
			}
		}
		if pos == 1 {
			bulletCount++
		}
	}

	if bulletCount != BulletCount {
		return &ValidationError{
			Field:   "Cartridges",
			Message: fmt.Sprintf("must have exactly %d bullet(s)", BulletCount),
			Value:   bulletCount,
		}
	}

	return nil
}

// 边界检查函数
// CheckBoundaries 检查数值边界
func CheckBoundaries(value, min, max int64, fieldName string) error {
	if value < min || value > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("value must be between %d and %d", min, max),
			Value:   value,
		}
	}
	return nil
}

// ValidateRange 验证范围
func ValidateRange(value int64, min, max int64, name string) error {
	if value < min {
		return &ValidationError{
			Field:   name,
			Message: fmt.Sprintf("value %d is below minimum %d", value, min),
			Value:   value,
		}
	}
	if value > max {
		return &ValidationError{
			Field:   name,
			Message: fmt.Sprintf("value %d is above maximum %d", value, max),
			Value:   value,
		}
	}
	return nil
}

// 集成验证
// ValidateAndSanitize 验证并清理数据
func ValidateAndSanitize(input interface{}) (interface{}, error) {
	switch v := input.(type) {
	case Session:
		validator := &SessionValidator{Session: v}
		if err := validator.Validate(); err != nil {
			return nil, err
		}
		return v, nil
	case int64:
		if v <= 0 {
			return nil, &ValidationError{
				Field:   "Value",
				Message: "value must be positive",
				Value:   v,
			}
		}
		return v, nil
	case string:
		if err := validateStringInput(v, 1, MaxMessageLength, "Input"); err != nil {
			return nil, err
		}
		return strings.TrimSpace(v), nil
	case []int:
		if len(v) == 0 {
			return nil, &ValidationError{
				Field:   "Array",
				Message: "array cannot be empty",
				Value:   v,
			}
		}
		return v, nil
	default:
		return nil, &ValidationError{
			Field:   "Type",
			Message: "unsupported type for validation",
			Value:   v,
		}
	}
}

// ValidateOnly 验证数据但不返回清理后的值（简化版本）
func ValidateOnly(input interface{}) error {
	_, err := ValidateAndSanitize(input)
	return err
}

// BatchValidate 批量验证
func BatchValidate(validators ...Validator) error {
	var errors []error

	for _, validator := range validators {
		if err := validator.Validate(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch validation failed with %d errors: %v", len(errors), errors)
	}

	return nil
}

// 预处理函数
// SanitizeInput 清理输入
func SanitizeInput(input string) string {
	// 去除首尾空格
	input = strings.TrimSpace(input)

	// 移除控制字符
	input = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return -1
		}
		return r
	}, input)

	// 转义HTML特殊字符（如果需要）
	// 这里可以添加HTML转义逻辑

	return input
}

// Normalize 归一化数据
func Normalize(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return strings.ToLower(strings.TrimSpace(v))
	case int64:
		if v < 0 {
			return 0
		}
		return v
	case []int:
		return unique(v)
	default:
		return v
	}
}
