// Package tgyj 同归于尽
package tgyj

import (
	"math/rand/v2"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	secondsPerMinute = 60
	minBanMinutes    = 1
	maxBanMinutes    = 3
	selfBanExtraMax  = 2
)

func init() {
	engine := control.Register("tgyj", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "同归于尽",
		Help:             "同归于尽@xxx",
	})
	engine.OnRegex(`同归于尽.*?(\d+)`, zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			targetUserID := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			targetBanMinutes := rand.IntN(maxBanMinutes-minBanMinutes+1) + minBanMinutes
			selfBanMinutes := targetBanMinutes + rand.IntN(selfBanExtraMax)

			banUser(ctx, targetUserID, targetBanMinutes)
			banUser(ctx, ctx.Event.UserID, selfBanMinutes)

			ctx.SendChain(message.At(ctx.Event.UserID),
				message.Text("\n你向"),
				message.At(targetUserID),
				message.Text("发动了同归于尽，对方获得"+strconv.Itoa(targetBanMinutes)+"分钟禁言，爆炸伤害波及到自己，自己获得"+strconv.Itoa(selfBanMinutes)+"分钟禁言"),
			)
		})
}

func banUser(ctx *zero.Ctx, userID int64, minutes int) {
	ctx.SetGroupBan(ctx.Event.GroupID, userID, int64(minutes)*secondsPerMinute)
}
