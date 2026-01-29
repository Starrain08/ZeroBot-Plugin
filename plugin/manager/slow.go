package manager

import (
	"sync"
	"time"

	"github.com/RomiChan/syncx"
	"github.com/fumiama/slowdo"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type slowSender struct {
	lazy     *syncx.Lazy[*slowdo.Job[*zero.Ctx, message.Segment]]
	lastUsed time.Time
}

var (
	slowsenders sync.Map
	cleanupOnce sync.Once
)

func init() {
	cleanupOnce.Do(func() {
		go cleanupUnusedSenders()
	})
}

func collectsend(ctx *zero.Ctx, msgs ...message.Segment) {
	id := ctx.Event.GroupID
	if id == 0 {
		return
	}
	value, _ := slowsenders.LoadOrStore(id, &slowSender{
		lazy: &syncx.Lazy[*slowdo.Job[*zero.Ctx, message.Segment]]{
			Init: func() *slowdo.Job[*zero.Ctx, message.Segment] {
				x, err := slowdo.NewJob(time.Second*5, ctx, func(ctx *zero.Ctx, msg []message.Segment) {
					if len(msg) == 1 {
						ctx.Send(msg)
						return
					}
					m := make(message.Message, len(msg))
					for i, item := range msg {
						m[i] = message.CustomNode(
							zero.BotConfig.NickName[0],
							ctx.Event.SelfID,
							message.Message{item})
					}
					ctx.SendGroupForwardMessage(id, m)
				})
				if err != nil {
					panic(err)
				}
				return x
			},
		},
		lastUsed: time.Now(),
	})
	ss := value.(*slowSender)
	ss.lastUsed = time.Now()
	job := ss.lazy.Get()
	for _, msg := range msgs {
		job.Add(msg)
	}
}

func cleanupUnusedSenders() {
	ticker := time.NewTicker(time.Hour * 6)
	defer ticker.Stop()

	for range ticker.C {
		slowsenders.Range(func(key, value any) bool {
			ss := value.(*slowSender)
			if ss == nil || ss.lazy == nil {
				slowsenders.Delete(key)
				return true
			}
			if time.Since(ss.lastUsed) > time.Hour*12 {
				slowsenders.Delete(key)
			}
			return true
		})
	}
}

func ShutdownSlowSenders() {
	slowsenders.Range(func(key, value any) bool {
		if value != nil {
			ss := value.(*slowSender)
			ss.lazy = nil
		}
		return true
	})
}
