package roulette

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"
)

// Session 会话操作
type Session struct {
	GroupID    int64   // 群id
	Creator    int64   // 创建者
	Users      []int64 // 参与者
	Max        int64   // 最大人数
	Cartridges []int   // 弹夹
	IsValid    bool    // 是否有效
	ExpireTime int64   // 过期时间
	CreateTime int64   // 创建时间
}

var fileLocks = make(map[string]*sync.RWMutex)
var locksMu sync.Mutex

func getFileLock(path string) *sync.RWMutex {
	locksMu.Lock()
	defer locksMu.Unlock()
	if lock, ok := fileLocks[path]; ok {
		return lock
	}
	lock := &sync.RWMutex{}
	fileLocks[path] = lock
	return lock
}

func checkFile(path string) {
	lock := getFileLock(path)
	lock.Lock()
	defer lock.Unlock()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err := os.Create(path)
		if err != nil {
			return
		}
	}
}

func saveItem(dataPath string, info Session) {
	lock := getFileLock(dataPath)
	lock.Lock()
	defer lock.Unlock()
	interact := loadSessionsSafe(dataPath)
	if len(interact) == 0 {
		interact = append(interact, info)
	} else {
		for i, v := range interact {
			if v.GroupID == info.GroupID {
				interact[i] = info
				break
			}
		}
	}
	bytes, err := json.Marshal(&interact)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(dataPath, bytes, 0644); err != nil {
		panic(err)
	}
}

func loadSessions(dataPath string) []Session {
	lock := getFileLock(dataPath)
	lock.RLock()
	defer lock.RUnlock()
	return loadSessionsSafe(dataPath)
}

func loadSessionsSafe(dataPath string) []Session {
	ss := make([]Session, 0)
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return ss
	}
	if err = json.Unmarshal(data, &ss); err != nil {
		return ss
	}
	return ss
}

func getSession(gid int64, dataPath string) *Session {
	interact := loadSessions(dataPath)
	for i := range interact {
		if interact[i].GroupID == gid {
			return &interact[i]
		}
	}
	return nil
}

func (cls *Session) setValid(isValid bool, dataPath string) {
	cls.IsValid = isValid
	saveItem(dataPath, *cls)
}

// 添加会话
func addSession(gid, uid int64, dataPath string) {
	cls := Session{}
	cls.GroupID = gid
	cls.Creator = uid
	cls.Users = append(cls.Users, uid)
	cls.IsValid = false
	cls.Max = 3
	cls.Cartridges = cls.rotateRoulette()
	cls.ExpireTime = 300
	cls.CreateTime = time.Now().Unix()

	saveItem(dataPath, cls)
}

// 获取参与人数
func (cls *Session) countUser() int {
	return len(cls.Users)
}

// 加入会话
func (cls *Session) addUser(userID int64, dataPath string) {
	cls.Users = append(cls.Users, userID)
	saveItem(dataPath, *cls)
}

// 关闭
func (cls *Session) close(dataPath string) {
	lock := getFileLock(dataPath)
	lock.Lock()
	defer lock.Unlock()
	interact := loadSessionsSafe(dataPath)

	run := make([]Session, 0)
	for _, v := range interact {
		if v.GroupID == cls.GroupID {
			continue
		}
		run = append(run, v)
	}

	bytes, err := json.Marshal(&run)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(dataPath, bytes, 0644); err != nil {
		panic(err)
	}
}

// 判断会话是否过期
func (cls *Session) isExpire() bool {
	now := time.Now().Unix()
	return cls.CreateTime+cls.ExpireTime < now
}

// 判断是否在队伍中
func (cls *Session) checkJoin(uid int64) bool {
	for _, j := range cls.Users {
		if j == uid {
			return true
		}
	}
	return false
}

// 判断是否轮到用户
func (cls *Session) checkTurn(uid int64) bool {
	return cls.Users[0] == uid
}

// 剩余子弹数
func (cls *Session) cartridgesLeft() int {
	return len(cls.Cartridges)
}

// 开火
func (cls *Session) openFire(dataPath string) bool {
	bullet := cls.Cartridges[0]
	cls.Cartridges = cls.Cartridges[1:]
	if bullet == 1 {
		return true
	}
	user := cls.Users[0]
	cls.Users = cls.Users[1:]
	cls.Users = append(cls.Users, user)

	saveItem(dataPath, *cls)
	return false
}

// 打乱参与人顺序
func (cls *Session) rotateUser(dataPath string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cls.Users), func(i, j int) { cls.Users[i], cls.Users[j] = cls.Users[j], cls.Users[i] })
	saveItem(dataPath, *cls)
}

// 旋转轮盘
func (cls *Session) rotateRoulette() []int {
	cartridges := []int{1, 0, 0, 0, 0, 0}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cartridges), func(i, j int) { cartridges[i], cartridges[j] = cartridges[j], cartridges[i] })
	return cartridges
}
