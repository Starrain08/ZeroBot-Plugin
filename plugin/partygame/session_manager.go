package partygame

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SessionManager 管理游戏会话，提供缓存和批量操作功能
type SessionManager struct {
	sessions     map[int64]Session
	sessionLock  ReadWriteLock
	dataPath     string
	cacheExpiry  time.Duration
	lastCleanup  time.Time
	asyncLocker  *AsyncLocker
	taskExecutor *ConcurrentTask
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager(dataPath string) *SessionManager {
	sm := &SessionManager{
		sessions:     make(map[int64]Session),
		dataPath:     dataPath,
		cacheExpiry:  5 * time.Minute,
		lastCleanup:  time.Now(),
		asyncLocker:  NewAsyncLocker(30 * time.Second),
		taskExecutor: NewConcurrentTask(5),
	}

	// 启动监控
	go sm.StartMonitoring(1 * time.Minute)

	return sm
}

// LoadAllSessions 从文件加载所有会话到缓存
func (sm *SessionManager) LoadAllSessions() error {
	sm.sessionLock.BeginWrite()
	defer sm.sessionLock.EndWrite()

	sessions, err := readAllSessionsFromFile(sm.dataPath)
	if err != nil {
		return err
	}

	sm.sessions = make(map[int64]Session)
	for _, session := range sessions {
		sm.sessions[session.GroupID] = session
	}

	logrus.Infof("[SessionManager]已加载 %d 个会话", len(sm.sessions))
	return nil
}

// SaveAllSessions 将所有会话保存到文件
func (sm *SessionManager) SaveAllSessions() error {
	sm.sessionLock.RLock()
	defer sm.sessionLock.RUnlock()

	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return writeAllSessionsToFile(sm.dataPath, sessions)
}

// GetSession 获取会话，带自动清理过期会话
func (sm *SessionManager) GetSession(groupID int64) (Session, error) {
	// 使用异步锁防止并发获取
	if !sm.asyncLocker.TryLock(strconv.FormatInt(groupID, 10)) {
		return Session{}, ErrSessionNotFound
	}
	defer sm.asyncLocker.Unlock(strconv.FormatInt(groupID, 10))

	sm.sessionLock.BeginRead()
	defer sm.sessionLock.EndRead()

	// 定期清理过期会话
	sm.cleanupExpiredSessions()

	session, exists := sm.sessions[groupID]
	if !exists {
		return Session{}, ErrSessionNotFound
	}

	if session.IsExpired() {
		delete(sm.sessions, groupID)
		return Session{}, ErrSessionExpired
	}

	return session, nil
}

// AddSession 添加会话到缓存
func (sm *SessionManager) AddSession(session Session) error {
	sm.sessionLock.BeginWrite()
	defer sm.sessionLock.EndWrite()

	sm.sessions[session.GroupID] = session
	return sm.SaveAllSessions()
}

// UpdateSession 更新会话
func (sm *SessionManager) UpdateSession(session Session) error {
	sm.sessionLock.BeginWrite()
	defer sm.sessionLock.EndWrite()

	if _, exists := sm.sessions[session.GroupID]; !exists {
		return ErrSessionNotFound
	}

	sm.sessions[session.GroupID] = session
	return sm.SaveAllSessions()
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(groupID int64) error {
	sm.sessionLock.BeginWrite()
	defer sm.sessionLock.EndWrite()

	if _, exists := sm.sessions[groupID]; !exists {
		return ErrSessionNotFound
	}

	delete(sm.sessions, groupID)
	return sm.SaveAllSessions()
}

// cleanupExpiredSessions 清理过期会话
func (sm *SessionManager) cleanupExpiredSessions() {
	if time.Since(sm.lastCleanup) < sm.cacheExpiry {
		return
	}

	expiredSessions := []int64{}
	for groupID, session := range sm.sessions {
		if session.IsExpired() {
			expiredSessions = append(expiredSessions, groupID)
		}
	}

	for _, groupID := range expiredSessions {
		delete(sm.sessions, groupID)
		logrus.Infof("[SessionManager]清理过期会话: groupID=%d", groupID)
	}

	sm.lastCleanup = time.Now()
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []Session {
	sm.sessionLock.BeginRead()
	defer sm.sessionLock.EndRead()

	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetActiveSessions 获取活跃会话
func (sm *SessionManager) GetActiveSessions() []Session {
	sm.sessionLock.BeginRead()
	defer sm.sessionLock.EndRead()

	sessions := make([]Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		if session.IsValid && !session.IsExpired() {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// BatchOperation 批量操作
func (sm *SessionManager) BatchOperation(operations []func() error) error {
	var operationErrors []error

	for _, op := range operations {
		if err := op(); err != nil {
			operationErrors = append(operationErrors, err)
		}
	}

	if len(operationErrors) > 0 {
		return fmt.Errorf("批量操作失败，错误数量: %d", len(operationErrors))
	}

	return nil
}

// 带缓存的文件读取优化
func readAllSessionsWithCache(path string, cache *SessionManager) ([]Session, error) {
	// 首先尝试从缓存读取
	if cache != nil {
		sessions := cache.GetAllSessions()
		if len(sessions) > 0 {
			return sessions, nil
		}
	}

	// 缓存未命中，从文件读取
	return readAllSessionsFromFile(path)
}

// 带缓存的文件写入优化
func writeAllSessionsWithCache(path string, sessions []Session, cache *SessionManager) error {
	// 更新缓存
	if cache != nil {
		for _, session := range sessions {
			if err := cache.AddSession(session); err != nil {
				logrus.Errorf("[SessionManager]更新缓存失败: %v", err)
			}
		}
	}

	// 写入文件
	return writeAllSessionsToFile(path, sessions)
}

// 原子性文件操作
func atomicWriteFile(path string, data []byte) error {
	// 创建临时文件
	tempPath := path + ".tmp"

	// 写入临时文件
	if err := os.WriteFile(tempPath, data, FilePermission); err != nil {
		return err
	}

	// 原子性重命名
	return os.Rename(tempPath, path)
}

// 会话状态监控
func (sm *SessionManager) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		sm.MonitorSessions()
	}
}

// MonitorSessions 监控会话状态
func (sm *SessionManager) MonitorSessions() {
	sm.sessionLock.RLock()
	defer sm.sessionLock.RUnlock()

	activeCount := 0
	expiredCount := 0
	totalCount := len(sm.sessions)

	for _, session := range sm.sessions {
		if session.IsExpired() {
			expiredCount++
		} else if session.IsValid {
			activeCount++
		}
	}

	logrus.Infof("[SessionManager]会话状态统计: 总数=%d, 活跃=%d, 过期=%d",
		totalCount, activeCount, expiredCount)
}

// 会话池管理
type SessionPool struct {
	pool    chan Session
	manager *SessionManager
	maxSize int
}

// NewSessionPool 创建会话池
func NewSessionPool(manager *SessionManager, maxSize int) *SessionPool {
	return &SessionPool{
		pool:    make(chan Session, maxSize),
		manager: manager,
		maxSize: maxSize,
	}
}

// GetSessionFromPool 从池中获取会话
func (sp *SessionPool) GetSessionFromPool(groupID int64) (Session, error) {
	select {
	case session := <-sp.pool:
		if session.GroupID == groupID && !session.IsExpired() {
			return session, nil
		}
		// 会话不匹配，放回池中
		sp.pool <- session
	default:
		// 池为空，从管理器获取
		return sp.manager.GetSession(groupID)
	}

	// 再次尝试从管理器获取
	return sp.manager.GetSession(groupID)
}

// ReturnSessionToPool 将会话返回池中
func (sp *SessionPool) ReturnSessionToPool(session Session) {
	select {
	case sp.pool <- session:
		// 成功返回池中
	default:
		// 池已满，不做处理
	}
}

// 获取会话统计信息
func (sm *SessionManager) GetStats() SessionStats {
	sm.sessionLock.RLock()
	defer sm.sessionLock.RUnlock()

	stats := SessionStats{
		TotalSessions:   len(sm.sessions),
		ActiveSessions:  0,
		ExpiredSessions: 0,
		TotalPlayers:    0,
	}

	for _, session := range sm.sessions {
		if session.IsExpired() {
			stats.ExpiredSessions++
		} else if session.IsValid {
			stats.ActiveSessions++
		}
		stats.TotalPlayers += len(session.Users)
	}

	return stats
}

// AsyncSave 异步保存会话
func (sm *SessionManager) AsyncSave(session Session) {
	sm.taskExecutor.AddTask(func() {
		if err := sm.UpdateSession(session); err != nil {
			logrus.Errorf("[SessionManager]异步保存会话失败: %v", err)
		}
	})
}

// SessionStats 会话统计信息
type SessionStats struct {
	TotalSessions   int `json:"total_sessions"`
	ActiveSessions  int `json:"active_sessions"`
	ExpiredSessions int `json:"expired_sessions"`
	TotalPlayers    int `json:"total_players"`
}

// 导出统计信息
func (stats SessionStats) String() string {
	return fmt.Sprintf(
		"会话统计 - 总数: %d, 活跃: %d, 过期: %d, 玩家总数: %d",
		stats.TotalSessions,
		stats.ActiveSessions,
		stats.ExpiredSessions,
		stats.TotalPlayers,
	)
}

// 文件I/O 性能优化
type FileCache struct {
	cache map[string][]byte
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewFileCache 创建文件缓存
func NewFileCache(ttl time.Duration) *FileCache {
	fc := &FileCache{
		cache: make(map[string][]byte),
		ttl:   ttl,
	}

	// 启动缓存清理
	go fc.startCleanup()
	return fc
}

// Get 从缓存获取文件内容
func (fc *FileCache) Get(path string) ([]byte, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	data, exists := fc.cache[path]
	if exists {
		// 检查是否过期
		// 这里可以添加过期逻辑
		return data, true
	}
	return nil, false
}

// Set 设置文件内容到缓存
func (fc *FileCache) Set(path string, data []byte) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.cache[path] = data
}

// Clear 清空缓存
func (fc *FileCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.cache = make(map[string][]byte)
}

// startCleanup 启动缓存清理
func (fc *FileCache) startCleanup() {
	ticker := time.NewTicker(fc.ttl)
	defer ticker.Stop()

	for range ticker.C {
		fc.cleanup()
	}
}

// cleanup 清理过期缓存
func (fc *FileCache) cleanup() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// 简单的清理策略：清空所有缓存
	// 实际应用中可以实现更复杂的LRU等策略
	fc.cache = make(map[string][]byte)

	logrus.Debugf("[FileCache]缓存已清理")
}
