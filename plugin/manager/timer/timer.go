package timer

import (
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/process"
	sql "github.com/FloatTech/sqlite"
	"github.com/fumiama/cron"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	maxTimerDuration = time.Hour * 24 * 365
)

// Clock 时钟
type Clock struct {
	db       *sql.Sqlite
	timers   map[uint32]*Timer
	timersmu sync.RWMutex
	cron     *cron.Cron
	entries  map[uint32]cron.EntryID
	entmu    sync.Mutex
}

var (
	// @全体成员
	atall = message.Segment{
		Type: "at",
		Data: map[string]string{
			"qq": "all",
		},
	}
)

// NewClock 添加一个新时钟
func NewClock(db *sql.Sqlite) (c *Clock) {
	c = &Clock{
		timers:  make(map[uint32]*Timer),
		cron:    cron.New(),
		entries: make(map[uint32]cron.EntryID),
	}
	c.loadTimers(db)
	c.cron.Start()
	return
}

// RegisterTimer 注册计时器
func (c *Clock) RegisterTimer(ts *Timer, save, isinit bool) bool {
	key := ts.ID
	if save {
		key = ts.GetTimerID()
		ts.ID = key
	}

	existingTimer, exists := c.GetTimer(key)
	if exists && existingTimer != ts {
		existingTimer.SetEn(false)
	}

	logrus.Infoln("[群管]注册计时器", key)

	if ts.Cron != "" {
		return c.registerCronTimer(ts, save, isinit)
	}
	return c.registerStandardTimer(ts, save)
}

func (c *Clock) registerCronTimer(ts *Timer, save, isinit bool) bool {
	ctx := c.getBotContext(ts, isinit)

	eid, err := c.cron.AddFunc(ts.Cron, func() {
		ts.sendmsg(ts.GrpID, ctx)
	})
	if err != nil {
		ts.Alert = err.Error()
		return false
	}

	c.entmu.Lock()
	c.entries[ts.ID] = eid
	c.entmu.Unlock()

	if save {
		if err := c.AddTimerIntoDB(ts); err != nil {
			return false
		}
	}
	return c.AddTimerIntoMap(ts) == nil
}

func (c *Clock) registerStandardTimer(ts *Timer, save bool) bool {
	if save {
		_ = c.AddTimerIntoDB(ts)
	}
	if err := c.AddTimerIntoMap(ts); err != nil {
		return false
	}

	go c.runStandardTimer(ts)
	return true
}

func (c *Clock) runStandardTimer(ts *Timer) {
	for ts.En() {
		nextdate := ts.nextWakeTime()
		if nextdate.Before(time.Now()) {
			nextdate = time.Now().Add(time.Minute)
		}

		duration := time.Until(nextdate)
		if duration > maxTimerDuration {
			logrus.Warnf("[群管]计时器%08x唤醒时间过远，跳过", ts.ID)
			continue
		}

		logrus.Printf("[群管]计时器%08x将睡眠%ds", ts.ID, duration/time.Second)

		if duration > 0 {
			time.Sleep(duration)
		}

		if !ts.En() {
			break
		}

		c.checkAndExecuteTimer(ts, nextdate)
	}
}

func (c *Clock) checkAndExecuteTimer(ts *Timer, nextdate time.Time) {
	now := time.Now()
	monthMatch := ts.Month() < 0 || ts.Month() == now.Month()
	dayMatch := ts.Day() < 0 || ts.Day() == now.Day()
	weekMatch := ts.Week() < 0 || (ts.Week() >= 0 && ts.Week() == now.Weekday())

	if monthMatch && (dayMatch || (ts.Day() == 0 && weekMatch)) {
		if ts.Hour() < 0 || ts.Hour() == now.Hour() {
			if ts.Minute() < 0 || ts.Minute() == now.Minute() {
				if ts.SelfID != 0 {
					ts.sendmsg(ts.GrpID, zero.GetBot(ts.SelfID))
				} else {
					zero.RangeBot(func(_ int64, ctx *zero.Ctx) bool {
						ts.sendmsg(ts.GrpID, ctx)
						return true
					})
				}
			}
		}
	}
}

func (c *Clock) getBotContext(ts *Timer, isinit bool) *zero.Ctx {
	if isinit {
		process.GlobalInitMutex.Lock()
		defer process.GlobalInitMutex.Unlock()
	}

	if ts.SelfID != 0 {
		return zero.GetBot(ts.SelfID)
	}

	var ctx *zero.Ctx
	zero.RangeBot(func(id int64, botCtx *zero.Ctx) bool {
		ctx = botCtx
		ts.SelfID = id
		return false
	})
	return ctx
}

// CancelTimer 取消计时器
func (c *Clock) CancelTimer(key uint32) bool {
	t, ok := c.GetTimer(key)
	if !ok {
		return false
	}

	t.SetEn(false)

	if t.Cron != "" {
		c.entmu.Lock()
		if eid, exists := c.entries[key]; exists {
			c.cron.Remove(eid)
			delete(c.entries, key)
		}
		c.entmu.Unlock()
	}

	c.timersmu.Lock()
	delete(c.timers, key)
	err := c.db.Del("timer", "WHERE id = ?", key)
	c.timersmu.Unlock()

	return err == nil
}

// ListTimers 列出本群所有计时器
func (c *Clock) ListTimers(grpID int64) []string {
	if c.timers == nil {
		return nil
	}

	c.timersmu.RLock()
	defer c.timersmu.RUnlock()

	keys := make([]string, 0, len(c.timers))
	for _, v := range c.timers {
		if v.GrpID == grpID {
			keys = append(keys, formatTimerInfo(v))
		}
	}
	return keys
}

func formatTimerInfo(t *Timer) string {
	k := t.GetTimerInfo()
	start := strings.Index(k, "]")
	if start == -1 {
		return k
	}
	msg := k[start+1:]
	msg = strings.ReplaceAll(msg+"\n", "-1", "每")
	msg = strings.ReplaceAll(msg, "月0日0周", "月周天")
	msg = strings.ReplaceAll(msg, "月0日", "月")
	msg = strings.ReplaceAll(msg, "日0周", "日")
	return msg
}

// GetTimer 获得定时器
func (c *Clock) GetTimer(key uint32) (t *Timer, ok bool) {
	c.timersmu.RLock()
	t, ok = c.timers[key]
	c.timersmu.RUnlock()
	return
}

// AddTimerIntoDB 添加定时器
func (c *Clock) AddTimerIntoDB(t *Timer) error {
	c.timersmu.Lock()
	defer c.timersmu.Unlock()
	return c.db.Insert("timer", t)
}

// AddTimerIntoMap 添加定时器到缓存
func (c *Clock) AddTimerIntoMap(t *Timer) error {
	c.timersmu.Lock()
	c.timers[t.ID] = t
	c.timersmu.Unlock()
	return nil
}

func (c *Clock) loadTimers(db *sql.Sqlite) {
	c.db = db
	err := c.db.Create("timer", &Timer{})
	if err != nil {
		return
	}

	var t Timer
	_ = c.db.FindFor("timer", &t, "", func() error {
		tescape := t
		go c.RegisterTimer(&tescape, false, true)
		return nil
	})
}

// Shutdown 优雅关闭时钟
func (c *Clock) Shutdown() {
	c.cron.Stop()
	c.timersmu.Lock()
	for _, t := range c.timers {
		t.SetEn(false)
	}
	c.timersmu.Unlock()
}
