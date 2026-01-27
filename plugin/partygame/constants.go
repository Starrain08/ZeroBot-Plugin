package partygame

// 游戏常量定义
const (
	// 玩家相关常量
	MaxPlayers         = 3          // 最大玩家数量
	MinPlayers         = 2          // 最小玩家数量
	DefaultCreatorID   = 1          // 默认创建者ID
	
	// 轮盘赌相关常量
	MaxCartridges      = 6          // 最大弹夹数量
	BulletCount        = 1          // 子弹数量
	CartridgeCapacity = 6          // 弹夹容量
	
	// 时间相关常量 (秒)
	TimeoutDuration    = 120        // 总超时时间
	WarningDuration    = 105        // 警告时间
	ResponseTimeout   = 20         // 回复超时时间
	SessionExpireTime  = 300        // 会话过期时间 (5分钟)
	
	// 文件相关常量
	FilePermission     = 0644       // 文件权限
	MaxFileSize       = 1024 * 1024 // 最大文件大小 (1MB)
	
	// 消息长度限制
	MaxMessageLength   = 500        // 最大消息长度
	MinNicknameLength  = 1          // 最小昵称长度
	MaxNicknameLength  = 20         // 最大昵称长度
	
	// 重试次数
	MaxRetryAttempts   = 3          // 最大重试次数
	
	// 缓存相关
	DefaultCacheSize   = 100        // 默认缓存大小
	CleanupInterval   = 300        // 清理间隔 (5分钟)
)

// 游戏状态常量
const (
	GameStateCreated   = "created"   // 游戏已创建
	GameStateActive    = "active"    // 游戏进行中
	GameStateFinished  = "finished"  // 游戏已结束
	GameStateExpired   = "expired"   // 游戏已过期
)

// 玩家角色常量
const (
	RoleCreator        = "creator"  // 创建者
	RolePlayer         = "player"   // 玩家
	RoleAdmin          = "admin"    // 管理员
	RoleSpectator      = "spectator" // 观众
)

// 消息类型常量
const (
	MsgTypeText        = "text"     // 文本消息
	MsgTypeAt          = "at"       // @消息
	MsgTypeImage       = "image"    // 图片消息
	MsgTypeFace        = "face"     // 表情消息
)

// 错误码常量
const (
	ErrCodeSuccess     = 0          // 成功
	ErrCodeInvalidParam = 1         // 无效参数
	ErrCodeNotFound    = 2         // 未找到
	ErrCodeExpired     = 3         // 已过期
	ErrCodePermission  = 4         // 权限不足
	ErrCodeInternal    = 5         // 内部错误
	ErrCodeTimeout     = 6         // 超时
	ErrCodeFull        = 7         // 已满
)

// 游戏规则常量
const (
	RouletteRule      = "轮盘容量为6, 但只填充了一发子弹, 请参与游戏的双方轮流发送`开火`, 枪响结束后"
	TruthOrDareRule    = "真心话大冒险，选择真心话或大冒险来接受惩罚"
	
	// 轮盘赌结果
	RouletteWin        = "win"      // 获胜
	RouletteLose      = "lose"     // 失败
)

// 配置常量
const (
	ConfigBotName      = "ZeroBot"  // 机器人默认名称
	ConfigLogLevel     = "info"     // 日志级别
	ConfigMaxConcurrent= 10         // 最大并发数
)

// 数据库相关常量
const (
	DbSessionsTable   = "sessions" // 会话表名
	DbUsersTable      = "users"    // 用户表名
	DbStatsTable      = "stats"    // 统计表名
)

// 统计指标常量
const (
	StatGamesStarted   = "games_started"   // 开始游戏次数
	StatGamesFinished  = "games_finished"   // 完成游戏次数
	StatPlayersJoined = "players_joined"   // 加入玩家次数
	StatTruthSelected = "truth_selected"   // 选择真心话次数
	StatDareSelected  = "dare_selected"    // 选择大冒险次数
)

// 正则表达式常量
const (
	RegexQQNumber      = `^\d{5,12}$`           // QQ号码正则
	RegexNickname      = `^[a-zA-Z0-9_\u4e00-\u9fa5]{1,20}$` // 昵称正则
	RegexTruthOrDare   = `^(真心话|大冒险)$`     // 真心话大冒险正则
	RegexRoulette      = `^(开火|终止轮盘赌)$`   // 轮盘赌正则
)

// API相关常量
const (
	ApiVersion         = "v1"               // API版本
	ApiRateLimit       = 60                 // API速率限制 (次/分钟)
	ApiTimeout         = 30                 // API超时时间 (秒)
)

// 缓存键前缀常量
const (
	CachePrefixSession  = "session:"          // 会话缓存前缀
	CachePrefixUser    = "user:"             // 用户缓存前缀
	CachePrefixGame    = "game:"             // 游戏缓存前缀
	CachePrefixStats   = "stats:"            // 统计缓存前缀
)

// 事件类型常量
const (
	EventTypeGameStart  = "game_start"       // 游戏开始事件
	EventTypeGameEnd    = "game_end"         // 游戏结束事件
	EventTypePlayerJoin = "player_join"      // 玩家加入事件
	EventTypePlayerLeave = "player_leave"    // 玩家离开事件
	EventTypeAction     = "action"           // 动作事件
)

// 优先级常量
const (
	PriorityHigh       = 3                 // 高优先级
	PriorityNormal     = 2                 // 普通优先级
	PriorityLow        = 1                 // 低优先级
)

// 重试策略常量
const (
	RetryPolicyLinear  = "linear"           // 线性重试
	RetryPolicyExponential = "exponential"  // 指数重试
	RetryPolicyFixed   = "fixed"            // 固定重试
)

// 健康检查常量
const (
	HealthCheckInterval = 30                // 健康检查间隔 (秒)
	HealthCheckTimeout  = 10                // 健康检查超时 (秒)
	MaxHealthFailures  = 3                 // 最大健康检查失败次数
)

// 监控指标常量
const (
	MetricResponseTime  = "response_time"    // 响应时间
	MetricErrorRate    = "error_rate"       // 错误率
	MetricThroughput   = "throughput"       // 吞吐量
	MetricActiveGames  = "active_games"     // 活跃游戏数
	MetricActiveUsers  = "active_users"     // 活跃用户数
)

// 安全相关常量
const (
	MaxLoginAttempts   = 5                 // 最大登录尝试次数
	LockoutDuration    = 300                // 锁定时间 (秒)
	SessionTokenLength = 32                 // 会话令牌长度
	MaxSessionTokens  = 5                  // 最大会话令牌数
)

// 限流相关常量
const (
	RateLimitWindow    = 60                 // 限流窗口 (秒)
	RateLimitCapacity  = 100                // 限流容量
	RateLimitBurst    = 10                 // 限流突发容量
)

// 日志相关常量
const (
	LogMaxSize         = 100                // 日志最大大小 (MB)
	LogMaxBackups      = 3                  // 日志最大备份数
	LogMaxAge          = 7                  // 日志最大保留天数
	LogCompress        = true               // 日志压缩
)