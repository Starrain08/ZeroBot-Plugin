package partygame

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	dataPath       string
	fileMu         sync.RWMutex
	sessionManager *SessionManager
	fileCache      *FileCache
)

type Session struct {
	GroupID    int64
	Creator    int64
	Users      []int64
	Max        int64
	Cartridges []int
	IsValid    bool
	ExpireTime int64
	CreateTime int64
}

func initializeDataPath(path string) error {
	if path == "" {
		return fmt.Errorf("数据路径不能为空")
	}

	fileMu.Lock()
	defer fileMu.Unlock()

	dataPath = path

	// 初始化文件缓存
	fileCache = NewFileCache(5 * time.Minute)
	
	// 初始化会话管理器
	sessionManager = NewSessionManager(path)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := atomicWriteFile(path, []byte("[]")); err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
	}

	// 预加载会话到缓存
	if err := sessionManager.LoadAllSessions(); err != nil {
		logrus.Warnf("[PartyGame]预加载会话失败: %v", err)
	}

	logrus.Infof("[PartyGame]数据路径初始化完成: %s", path)
	return nil
}

func readAllSessionsFromFile(path string) ([]Session, error) {
	// 尝试从缓存获取
	if fileCache != nil {
		if data, exists := fileCache.Get(path); exists {
			var sessions []Session
			if err := json.Unmarshal(data, &sessions); err == nil {
				return sessions, nil
			}
		}
	}

	fileMu.RLock()
	defer fileMu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	if len(data) == 0 {
		return []Session{}, nil
	}

	// 缓存文件内容
	if fileCache != nil {
		fileCache.Set(path, data)
	}

	var sessions []Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}
	return sessions, nil
}

func writeAllSessionsToFile(path string, sessions []Session) error {
	bytes, err := json.Marshal(&sessions)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// 原子性写入
	if err := atomicWriteFile(path, bytes); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 更新缓存
	if fileCache != nil {
		fileCache.Set(path, bytes)
	}

	logrus.Debugf("[PartyGame]会话数据已保存，数量: %d", len(sessions))
	return nil
}

func saveSession(path string, session Session) error {
	// 优先使用会话管理器
	if sessionManager != nil {
		return sessionManager.UpdateSession(session)
	}

	// 回退到文件操作
	sessions, err := readAllSessionsFromFile(path)
	if err != nil {
		return err
	}

	updated := false
	for i, s := range sessions {
		if s.GroupID == session.GroupID {
			sessions[i] = session
			updated = true
			break
		}
	}

	if !updated {
		sessions = append(sessions, session)
	}

	return writeAllSessionsToFile(path, sessions)
}

func findSessionByGroupID(path string, groupID int64) (Session, error) {
	// 优先使用会话管理器
	if sessionManager != nil {
		return sessionManager.GetSession(groupID)
	}

	// 回退到文件操作
	sessions, err := readAllSessionsFromFile(path)
	if err != nil {
		return Session{}, err
	}

	for _, s := range sessions {
		if s.GroupID == groupID {
			return s, nil
		}
	}
	return Session{}, ErrSessionNotFound
}

func createNewSession(path string, groupID, creatorID int64) error {
	// 验证输入参数
	if err := ValidateAndSanitize(groupID); err != nil {
		return fmt.Errorf("无效的群组ID: %w", err)
	}
	
	if err := ValidateAndSanitize(creatorID); err != nil {
		return fmt.Errorf("无效的创建者ID: %w", err)
	}

	session := Session{
		GroupID:    groupID,
		Creator:    creatorID,
		Users:      []int64{creatorID},
		IsValid:    false,
		Max:        MaxPlayers,
		Cartridges: generateRouletteCartridges(),
		ExpireTime: SessionExpireTime,
		CreateTime: time.Now().Unix(),
	}
	
	// 验证会话数据
	if err := (&SessionValidator{Session: session}).Validate(); err != nil {
		return fmt.Errorf("会话数据验证失败: %w", err)
	}
	return saveSession(path, session)
}

func deleteSession(path string, groupID int64) error {
	// 优先使用会话管理器
	if sessionManager != nil {
		return sessionManager.DeleteSession(groupID)
	}

	// 回退到文件操作
	sessions, err := readAllSessionsFromFile(path)
	if err != nil {
		return err
	}

	filtered := make([]Session, 0, len(sessions))
	for _, s := range sessions {
		if s.GroupID != groupID {
			filtered = append(filtered, s)
		}
	}

	return writeAllSessionsToFile(path, filtered)
}

func (s Session) GetUserCount() int {
	return len(s.Users)
}

func (s Session) GetMaxPlayers() int {
	return int(s.Max)
}

func (s Session) GetMinPlayers() int {
	return MinPlayers
}

func (s Session) GetGameState() string {
	if s.IsExpired() {
		return GameStateExpired
	}
	if s.IsValid {
		return GameStateActive
	}
	return GameStateCreated
}

// Validate 验证会话数据
func (s Session) Validate() error {
	validator := &SessionValidator{Session: s}
	return validator.Validate()
}

func (s *Session) AddPlayer(userID int64) error {
	// 验证用户ID
	if err := ValidateAndSanitize(userID); err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}

	if s.IsPlayerInSession(userID) {
		return ErrPlayerInSession
	}

	if int(s.Max) <= len(s.Users) {
		return fmt.Errorf("达到最大玩家数量 %d", s.Max)
	}

	s.Users = append(s.Users, userID)
	
	// 验证会话数据
	if err := s.Validate(); err != nil {
		return fmt.Errorf("会话数据验证失败: %w", err)
	}
	
	return saveSession(dataPath, *s)
}

func (s *Session) Terminate() error {
	logrus.Infof("[PartyGame]终止游戏会话: groupID=%d, players=%d", s.GroupID, len(s.Users))
	return deleteSession(dataPath, s.GroupID)
}

func (s Session) IsExpired() bool {
	return s.CreateTime+s.ExpireTime < time.Now().Unix()
}

func (s Session) IsPlayerInSession(userID int64) bool {
	for _, u := range s.Users {
		if u == userID {
			return true
		}
	}
	return false
}

func (s Session) IsPlayerTurn(userID int64) bool {
	if len(s.Users) == 0 {
		return false
	}
	return s.Users[0] == userID
}

func (s Session) GetRemainingCartridges() int {
	return len(s.Cartridges)
}

func (s *Session) ShufflePlayersOrder() error {
	if len(s.Users) < 2 {
		return nil
	}

	players := make([]int64, len(s.Users))
	copy(players, s.Users)

	randSource.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})

	s.Users = players
	
	// 异步保存
	if sessionManager != nil {
		sessionManager.AsyncSave(*s)
		return nil
	}
	
	return saveSession(dataPath, *s)
}

func generateRouletteCartridges() []int {
	cartridges := make([]int, CartridgeCapacity)
	// 填充子弹
	for i := 0; i < BulletCount; i++ {
		cartridges[i] = 1
	}
	// 其余位置填充空弹
	for i := BulletCount; i < CartridgeCapacity; i++ {
		cartridges[i] = 0
	}
	
	// 打乱顺序
	randSource.Shuffle(CartridgeCapacity, func(i, j int) {
		cartridges[i], cartridges[j] = cartridges[j], cartridges[i]
	})
	
	// 验证生成的弹夹
	if err := ValidateCartridges(cartridges); err != nil {
		logrus.Errorf("[PartyGame]弹夹生成失败: %v", err)
		// 返回默认配置
		return []int{1, 0, 0, 0, 0, 0}
	}
	
	return cartridges
}

func (s *Session) Fire() (bool, error) {
	if len(s.Cartridges) == 0 {
		return false, ErrCartridgeEmpty
	}

	// 验证弹夹数据
	if err := ValidateCartridges(s.Cartridges); err != nil {
		logrus.Errorf("[PartyGame]弹夹数据验证失败: %v", err)
		return false, fmt.Errorf("弹夹数据异常")
	}

	bullet := s.Cartridges[0]
	s.Cartridges = s.Cartridges[1:]

	if bullet == 1 {
		return true, nil
	}

	if len(s.Users) < MinPlayers {
		return false, fmt.Errorf("玩家不足，至少需要%d名玩家", MinPlayers)
	}

	currentPlayer := s.Users[0]
	s.Users = s.Users[1:]
	s.Users = append(s.Users, currentPlayer)

	// 验证会话数据
	if err := s.Validate(); err != nil {
		return false, fmt.Errorf("会话数据验证失败: %w", err)
	}

	// 异步保存
	if sessionManager != nil {
		sessionManager.AsyncSave(*s)
		return false, nil
	}
	
	return false, saveSession(dataPath, *s)
}
