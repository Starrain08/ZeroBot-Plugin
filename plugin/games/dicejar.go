// Package gamesystem 基于zbp的猜歌插件
package gamesystem

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/games/gamesystem"
)

func init() {
	engine, gameManager, err := gamesystem.Register("骰子壶", &gamesystem.GameInfo{
		Command: "- 骰子壶@对方QQ",
		Help:    "保证星之尘大于20的双方玩家各自投掷1个骰子。\n平局的场合再掷一次直到数目不一样。",
		Rewards: "投掷出来的数目低的玩家将另一方投掷出的数目x2的星之尘交给对方。\n" +
			"如果输给投掷出来的数目为6的场合,移交的星之尘变成20。",
	})
	if err != nil {
		panic(err)
	}

	engine.OnRegex(`^骰子壶\s*?\[CQ:at,qq=(\d+).*`, zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		if gameManager.PlayIn(ctx.Event.GroupID) {
			return true
		}
		ctx.SendChain(message.Text("游戏已下架,无法游玩"))
		return false
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		matches := ctx.State["regex_matched"].([]string)
		if len(matches) < 2 {
			ctx.SendChain(message.Text("[ERROR]: 解析对方QQ失败"))
			return
		}

		duelUser, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]: QQ号码格式错误"))
			return
		}

		uid := ctx.Event.UserID
		if duelUser == uid {
			ctx.SendChain(message.Text("禁止左右手互博"))
			return
		}

		userBalance := wallet.GetWalletOf(uid)
		if userBalance < 20 {
			ctx.SendChain(message.Text("你的星之尘不足以满足该游戏"))
			return
		}

		challengerBalance := wallet.GetWalletOf(duelUser)
		if challengerBalance < 20 {
			ctx.SendChain(message.Text("他的星之尘不足以满足该游戏"))
			return
		}

		ctx.SendChain(message.Text("等待对方发送\"开玩|拒绝\"进行回复"))

		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.FullMatchRule("开玩", "拒绝"), zero.CheckUser(duelUser), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()

		timeout := time.NewTimer(120 * time.Second)
		defer timeout.Stop()

		answer := ""
		answerReceived := make(chan string, 1)

		go func() {
			for c := range recv {
				answerReceived <- c.Event.Message.String()
			}
		}()

		select {
		case <-timeout.C:
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("时间超时,游戏取消")))
			return
		case ans := <-answerReceived:
			answer = ans
		}

		if answer == "拒绝" {
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("对方拒绝了你的邀请,游戏结束")))
			return
		}

		if answer != "开玩" {
			return
		}

		duel := make(map[int64]int)
		wait := time.NewTimer(120 * time.Second)
		defer wait.Stop()

		ctx.SendChain(message.Text("游戏开始,请说出你的带\"投\"或\"骰\"的话语进行投掷骰子"))

		recv, cancel = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.KeywordRule("投", "骰"), zero.CheckUser(duelUser, uid), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()

		diceChan := make(chan struct {
			uid  int64
			dice int
		}, 2)

		go func() {
			for c := range recv {
				diceChan <- struct {
					uid  int64
					dice int
				}{c.Event.UserID, rand.Intn(6) + 1}
			}
		}()

		for {
			select {
			case <-wait.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("等待超时,游戏取消\n咕之人扣除 6 星之尘")))
				if _, ok := duel[uid]; !ok {
					err := wallet.InsertWalletOf(uid, -6)
					if err != nil {
						ctx.SendChain(message.Text("ERROR]:", err))
					}
				}
				if _, ok := duel[duelUser]; !ok {
					err := wallet.InsertWalletOf(duelUser, -6)
					if err != nil {
						ctx.SendChain(message.Text("ERROR]:", err))
					}
				}
				return
			case roll := <-diceChan:
				if _, ok := duel[roll.uid]; !ok {
					duel[roll.uid] = roll.dice
					ctx.SendChain(message.At(roll.uid), message.Text("\n你投掷出了数目:", roll.dice))
				}

				dice1, ok1 := duel[uid]
				dice2, ok2 := duel[duelUser]

				if ok1 && ok2 {
					if dice1 != dice2 {
						goto endLoop
					}
					ctx.SendChain(message.Text("你们投掷的数目相同,请重新投掷"))
					duel = make(map[int64]int)
				}
			}
		}

	endLoop:
		cancel()

		result := duel[uid] - duel[duelUser]
		points := 20
		switch {
		case result > 0:
			if duel[uid] != 6 {
				points = duel[uid] * 2
			}
		case result < 0:
			if duel[duelUser] != 6 {
				points = duel[duelUser] * 2
			}
			uid = duelUser
			duelUser = ctx.Event.UserID
		}

		err = wallet.InsertWalletOf(uid, points)
		if err == nil {
			err = wallet.InsertWalletOf(duelUser, -points)
		}

		if err != nil {
			ctx.SendChain(message.Text("ERROR]:", err))
			return
		}

		ctx.SendChain(message.At(uid), message.Text("恭喜你赢得了「骰子壶」游戏,获得"), message.At(duelUser), message.Text("的 ", points, " 星之尘"))
	})
}
