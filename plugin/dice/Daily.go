package dice

import (
	"strconv"
	"strings"

	fcext "github.com/FloatTech/floatbox/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex(`^[。.]jrrp`).SetBlock(true).
		Handle(handleJRRP)
	engine.OnRegex(`^设置jrrp([\s\S]*)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(handleSetJRRP)
}

func handleJRRP(ctx *zero.Ctx) {
	uid := ctx.Event.UserID
	jrrp := fcext.RandSenderPerDayN(uid, 100) + 1

	var j strjrrp
	if err := db.Find("strjrrp", &j, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10)); err == nil {
		ctx.SendGroupMessage(ctx.Event.GroupID, formatCustomJRRP(ctx, j.Strjrrp))
	} else {
		ctx.SendChain(message.At(uid), message.Text("阁下今日的人品值为", jrrp, "呢~"))
	}
}

func handleSetJRRP(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	j := &strjrrp{
		GrpID:   ctx.Event.GroupID,
		Strjrrp: matches[1],
	}
	if err := db.Insert("strjrrp", j); err == nil {
		ctx.SendChain(message.Text("记住啦!"))
	} else {
		ctx.SendChain(message.Text("ERROR: ", err))
	}
}

func formatCustomJRRP(ctx *zero.Ctx, template string) string {
	uid := strconv.FormatInt(ctx.Event.UserID, 10)
	at := "[CQ:at,qq=" + uid + "]"
	jrrp := fcext.RandSenderPerDayN(ctx.Event.UserID, 100) + 1
	jrrpStr := strconv.Itoa(jrrp)

	result := strings.ReplaceAll(template, "{jrrp}", jrrpStr)
	result = strings.ReplaceAll(result, "{at}", at)
	return result
}
