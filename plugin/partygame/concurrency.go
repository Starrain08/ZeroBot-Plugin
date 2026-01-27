package partygame

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"context"
)

// AsyncLocker 异步锁管理器
type AsyncLocker struct {
	locks    map[string]*sync.Mutex
	mu       sync.RWMutex
	timeout  time.Duration
	cleanup  *time.Ticker
}

// NewAsyncLocker 创建新的异步锁管理器
func NewAsyncLocker(timeout time.Duration) *AsyncLocker {
	al := &AsyncLocker{
		locks:   make(map[string]*sync.Mutex),
		timeout: timeout,
		cleanup: time.NewTicker(5 * time.Minute),
	}
	
	go al.cleanupExpiredLocks()
	return al
}

// TryLock 尝试获取锁
func (al *AsyncLocker) TryLock(key string) bool {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	if lock, exists := al.locks[key]; exists {
		return false
	}
	
	lock := &sync.Mutex{}
	lock.Lock()
	al.locks[key] = lock
	return true
}

// Unlock 释放锁
func (al *AsyncLocker) Unlock(key string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	if lock, exists := al.locks[key]; exists {
		lock.Unlock()
		delete(al.locks, key)
	}
}

// cleanupExpiredLocks 清理过期锁
func (al *AsyncLocker) cleanupExpiredLocks() {
	for range al.cleanup.C {
		al.cleanupExpired()
	}
}

func (al *AsyncLocker) cleanupExpired() {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// 这里可以添加更复杂的过期逻辑
	// 目前简单清理所有锁
	for key, lock := range al.locks {
		lock.Unlock()
		delete(al.locks, key)
	}
	
	logrus.Debugf("[AsyncLocker]清理过期锁，剩余: %d", len(al.locks))
}

// ConcurrentTask 并发任务执行器
type ConcurrentTask struct {
	maxWorkers int
	taskQueue  chan func()
	wg         sync.WaitGroup
}

// NewConcurrentTask 创建新的并发任务执行器
func NewConcurrentTask(maxWorkers int) *ConcurrentTask {
	ct := &ConcurrentTask{
		maxWorkers: maxWorkers,
		taskQueue:  make(chan func(), 100),
	}
	
	ct.startWorkers()
	return ct
}

// AddTask 添加任务
func (ct *ConcurrentTask) AddTask(task func()) {
	ct.taskQueue <- task
}

// startWorkers 启动工作线程
func (ct *ConcurrentTask) startWorkers() {
	for i := 0; i < ct.maxWorkers; i++ {
		ct.wg.Add(1)
		go ct.worker()
	}
}

// worker 工作线程
func (ct *ConcurrentTask) worker() {
	defer ct.wg.Done()
	
	for task := range ct.taskQueue {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("[ConcurrentTask]任务执行失败: %v", r)
				}
			}()
			task()
		}()
	}
}

// Wait 等待所有任务完成
func (ct *ConcurrentTask) Wait() {
	close(ct.taskQueue)
	ct.wg.Wait()
}

// ReadWriteLock 读写锁优化
type ReadWriteLock struct {
	readers int
	writer  bool
	readMu  sync.Mutex
	writeMu sync.Mutex
}

// LockRead 获取读锁
func (rw *ReadWriteLock) LockRead() {
	rw.readMu.Lock()
	if rw.readers == 0 {
		rw.writeMu.Lock()
	}
	rw.readers++
	rw.readMu.Unlock()
}

// UnlockRead 释放读锁
func (rw *ReadWriteLock) UnlockRead() {
	rw.readMu.Lock()
	rw.readers--
	if rw.readers == 0 {
		rw.writeMu.Unlock()
	}
	rw.readMu.Unlock()
}

// LockWrite 获取写锁
func (rw *ReadWriteLock) LockWrite() {
	rw.writeMu.Lock()
}

// UnlockWrite 释放写锁
func (rw *ReadWriteLock) UnlockWrite() {
	rw.writeMu.Unlock()
}

// SpinLock 自旋锁实现
type SpinLock struct {
	state int32
}

// Lock 获取自旋锁
func (sl *SpinLock) Lock() {
	for {
		if atomic.CompareAndSwapInt32(&sl.state, 0, 1) {
			return
		}
		runtime.Gosched()
	}
}

// Unlock 释放自旋锁
func (sl *SpinLock) Unlock() {
	atomic.StoreInt32(&sl.state, 0)
}

// OptimizedMutex 优化的互斥锁
type OptimizedMutex struct {
	initialized bool
	mu          sync.Mutex
	spinLock    SpinLock
	useSpinLock bool
}

// NewOptimizedMutex 创建优化的互斥锁
func NewOptimizedMutex() *OptimizedMutex {
	return &OptimizedMutex{
		spinLock: SpinLock{},
	}
}

// Lock 获取锁
func (om *OptimizedMutex) Lock() {
	if !om.initialized {
		om.mu.Lock()
		om.initialized = true
		om.mu.Unlock()
	}
	
	// 短时间竞争使用自旋锁
	if om.useSpinLock {
		om.spinLock.Lock()
	} else {
		om.mu.Lock()
	}
}

// Unlock 释放锁
func (om *OptimizedMutex) Unlock() {
	if om.useSpinLock {
		om.spinLock.Unlock()
	} else {
		om.mu.Unlock()
	}
}

// LazyWriteLazyRead 惰性写入惰性读取锁
type LazyWriteLazyRead struct {
	mu      sync.Mutex
	readers int
	writer  bool
}

// BeginRead 开始读取
func (lwlr *LazyWriteLazyRead) BeginRead() {
	lwlr.mu.Lock()
	if lwlr.readers == 0 {
		lwlr.writer = true
	}
	lwlr.readers++
	lwlr.mu.Unlock()
}

// EndRead 结束读取
func (lwlr *LazyWriteLazyRead) EndRead() {
	lwlr.mu.Lock()
	lwlr.readers--
	if lwlr.readers == 0 {
		lwlr.writer = false
	}
	lwlr.mu.Unlock()
}

// BeginWrite 开始写入
func (lwlr *LazyWriteLazyRead) BeginWrite() {
	lwlr.mu.Lock()
	for lwlr.readers > 0 || lwlr.writer {
		lwlr.mu.Unlock()
		lwlr.mu.Lock()
	}
	lwlr.writer = true
	lwlr.mu.Unlock()
}

// EndWrite 结束写入
func (lwlr *LazyWriteLazyRead) EndWrite() {
	lwlr.mu.Lock()
	lwlr.writer = false
	lwlr.mu.Unlock()
}

// Pool 连接池实现
type Pool struct {
	items   chan interface{}
	factory func() interface{}
	mu      sync.Mutex
}

// NewPool 创建连接池
func NewPool(size int, factory func() interface{}) *Pool {
	pool := &Pool{
		items:   make(chan interface{}, size),
		factory: factory,
	}
	
	// 预填充池
	for i := 0; i < size; i++ {
		pool.items <- pool.factory()
	}
	
	return pool
}

// Get 获取项目
func (p *Pool) Get() interface{} {
	select {
	case item := <-p.items:
		return item
	default:
		return p.factory()
	}
}

// Put 归还项目
func (p *Pool) Put(item interface{}) {
	select {
	case p.items <- item:
		// 成功归还
	default:
		// 池已满，丢弃
	}
}

// Close 关闭池
func (p *Pool) Close() {
	close(p.items)
}

// 条件变量优化
type OptimizedCondition struct {
	notify chan struct{}
	mu     sync.Mutex
}

// NewOptimizedCondition 创建优化的条件变量
func NewOptimizedCondition() *OptimizedCondition {
	return &OptimizedCondition{
		notify: make(chan struct{}, 1),
	}
}

// Wait 等待
func (oc *OptimizedCondition) Wait() {
	oc.mu.Lock()
	oc.mu.Unlock()
	
	<-oc.notify
}

// Signal 发送信号
func (oc *OptimizedCondition) Signal() {
	select {
	case oc.notify <- struct{}{}:
	default:
		// 信号已经存在
	}
}

// Broadcast 广播信号
func (oc *OptimizedCondition) Broadcast() {
	// 实现广播逻辑
}

// 超时锁
type TimeoutLock struct {
.mu    sync.Mutex
.owner string
}

// TryLockWithTimeout 尝试获取带超时的锁
func (tl *TimeoutLock) TryLockWithTimeout(owner string, timeout time.Duration) bool {
	start := time.Now()
	
	for {
		tl.mu.Lock()
		if tl.owner == "" {
			tl.owner = owner
			tl.mu.Unlock()
			return true
		}
		tl.mu.Unlock()
		
		if time.Since(start) > timeout {
			return false
		}
		
		time.Sleep(10 * time.Millisecond)
	}
}

// Unlock 释放锁
func (tl *TimeoutLock) Unlock(owner string) bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	if tl.owner == owner {
		tl.owner = ""
		return true
	}
	return false
}

// 分布式锁模拟
type DistributedLock struct {
	key      string
	owner    string
	ttl      time.Duration
	locked   bool
	lockChan chan struct{}
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		key:      key,
		ttl:      ttl,
		lockChan: make(chan struct{}, 1),
	}
}

// TryLock 尝试获取分布式锁
func (dl *DistributedLock) TryLock(owner string) bool {
	// 简化的分布式锁实现
	// 实际应用中应该使用Redis等外部存储
	
	select {
	case dl.lockChan <- struct{}{}:
		dl.owner = owner
		dl.locked = true
		return true
	default:
		return false
	}
}

// Unlock 释放分布式锁
func (dl *DistributedLock) Unlock(owner string) bool {
	if dl.locked && dl.owner == owner {
		<-dl.lockChan
		dl.locked = false
		dl.owner = ""
		return true
	}
	return false
}

// IsLocked 检查锁是否被持有
func (dl *DistributedLock) IsLocked() bool {
	return dl.locked
}

// GetOwner 获取锁所有者
func (dl *DistributedLock) GetOwner() string {
	return dl.owner
}

// 原子操作优化
type AtomicCounter struct {
	value int64
	mu    sync.RWMutex
}

// Increment 原子递增
func (ac *AtomicCounter) Increment() int64 {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.value++
	return ac.value
}

// Decrement 原子递减
func (ac *AtomicCounter) Decrement() int64 {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.value--
	return ac.value
}

// Get 原子获取值
func (ac *AtomicCounter) Get() int64 {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.value
}

// Set 原子设置值
func (ac *AtomicCounter) Set(value int64) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.value = value
}