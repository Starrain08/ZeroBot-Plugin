// Package gamesystem 基于zbp的猜歌插件
package gamesystem

import (
	"math/rand"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/games/gamesystem"
)

const (
	rscRock         = 1
	rscScissors     = 2
	rscPaper        = 3
	rscRewardChance = 5
	rscMaxReward    = 10
)

var point = map[string]int{
	"石头": rscRock,
	"剪刀": rscScissors,
	"布":  rscPaper,
}

func init() {
	engine, gameManager, err := gamesystem.Register("石头剪刀布", &gamesystem.GameInfo{
		Command: "- @bot[石头｜剪刀｜布]",
		Help:    "和机器人进行猜拳,如果机器人开心了会得到星之尘",
		Rewards: "奖励范围在0~10之间",
	})
	if err != nil {
		panic(err)
	}

	engine.OnFullMatchGroup([]string{"石头", "剪刀", "布"}, zero.OnlyToMe, func(ctx *zero.Ctx) bool {
		if gameManager.PlayIn(ctx.Event.GroupID) {
			return true
		}
		ctx.SendChain(message.Text("游戏已下架,无法游玩"))
		return false
	}).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(handleRockPaperScissors)
}

func handleRockPaperScissors(ctx *zero.Ctx) {
	botChoice := rscRock
	randIndex := rand.Intn(3)
	switch randIndex {
	case 0:
		botChoice = rscRock
	case 1:
		botChoice = rscScissors
	case 2:
		botChoice = rscPaper
	}

	switch botChoice {
	case rscRock:
		ctx.SendChain(message.Text("石头"))
	case rscScissors:
		ctx.SendChain(message.Text("剪刀"))
	case rscPaper:
		ctx.SendChain(message.Text("布"))
	}

	model := ctx.State["matched"].(string)
	playerChoice, ok := point[model]
	if !ok {
		return
	}

	result := playerChoice - botChoice

	if math.Abs(result) == 2 {
		result = -result
	}

	switch {
	case result < 0:
		ctx.SendChain(message.Text("可恶,你赢了"))
	case result > 0:
		if rand.Intn(rscRewardChance) == 1 {
			money := rand.Intn(rscMaxReward + 1)
			if money > 0 {
				err := wallet.InsertWalletOf(ctx.Event.UserID, money)
				if err == nil {
					ctx.SendChain(message.Text("哈哈,你输了。嗯!~今天运气不错,我很高兴,奖励你 ", money, " 枚星之尘吧"))
					return
				}
			}
		}
		ctx.SendChain(message.Text("哈哈,你输了"))
	default:
		if rand.Intn(10) == 1 {
			money := rand.Intn(rscMaxReward + 1)
			if money > 0 {
				err := wallet.InsertWalletOf(ctx.Event.UserID, money)
				if err == nil {
					ctx.SendChain(message.Text("你实力不错,我很欣赏你,奖励你 ", money, " 枚星之尘吧"))
					return
				}
			}
		}
		ctx.SendChain(message.Text("实力可以啊,希望下次再来和我玩"))
	}
}
