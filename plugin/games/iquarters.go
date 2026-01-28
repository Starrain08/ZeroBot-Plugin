// Package gamesystem 基于zbp的猜硬币游戏
package gamesystem

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/games/gamesystem"
)

const (
	iquartersMinScore    = 10
	iquartersMultiplier1 = 10
	iquartersMultiplier2 = 5
	iquartersPenalty     = 10
	iquartersTimeout     = 120 * time.Second
)

type iquartersGame struct {
	sessions map[int64]bool
	mu       sync.Mutex
}

var iquartersInstance = &iquartersGame{
	sessions: make(map[int64]bool),
}

func init() {
	engine, gameManager, err := gamesystem.Register("猜硬币", &gamesystem.GameInfo{
		Command: "- 创建猜硬币\n" +
			"- [加入|开始]游戏\n" +
			"- 开始投币",
		Help: "每个人宣言银币正面数量后,掷出参游人数的银币",
		Rewards: "正面与宣言的数量相同的场合获得 正面数*10 星之尘\n" +
			"正面与宣言的数量相差2以内的场合获得 正面数*5 星之尘\n" +
			"其他的的场合失去 10 星之尘",
	})
	if err != nil {
		panic(err)
	}

	engine.OnFullMatch("创建猜硬币", zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		id := ctx.Event.GroupID
		if iquartersInstance.isSessionActive(id) {
			ctx.SendChain(message.Text("游戏已在进行中"))
			return false
		}

		err := gameManager.CreateRoom(id)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	}).SetBlock(true).Handle(iquartersInstance.handleCreateGame)

	engine.OnFullMatch("加入游戏", zero.OnlyGroup).Handle(iquartersInstance.handleJoinGame)
	engine.OnFullMatch("开始游戏", zero.OnlyGroup).Handle(iquartersInstance.handleStartGame)
}

func (g *iquartersGame) isSessionActive(groupID int64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.sessions[groupID]
}

func (g *iquartersGame) handleCreateGame(ctx *zero.Ctx) {
	defer g.endGame(ctx.Event.GroupID)

	uid := ctx.Event.UserID
	if wallet.GetWalletOf(uid) < iquartersMinScore {
		ctx.SendChain(message.Text("你的星之尘不足以满足该游戏"))
		return
	}

	ctx.SendChain(message.Text("你开启了猜硬币游戏。\n其他人可发送\"加入游戏\"加入游戏或你发送\"开始游戏\"开始游戏"))

	recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.FullMatchRule("加入游戏", "开始游戏"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
	defer cancel()

	duel := make(map[int64]int)
	uidlist := []int64{uid}
	duel[uid] = -1

	timeout := time.NewTimer(iquartersTimeout)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("时间超时,游戏取消")))
			return
		case c := <-recv:
			answer := c.Event.Message.String()
			answerid := c.Event.UserID

			if answer == "加入游戏" {
				if _, ok := duel[answerid]; ok {
					ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
						message.Text("你已经加入了游戏")))
				} else {
					if wallet.GetWalletOf(answerid) < iquartersMinScore {
						ctx.SendChain(message.Text("你的星之尘不足以满足该游戏"))
						return
					}
					duel[answerid] = -1
					uidlist = append(uidlist, answerid)
					ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
						message.Text("成功加入游戏,等待开房人开始游戏")))
				}
			}

			if answer == "开始游戏" && answerid == uid {
				break
			}
		}
	}

	g.mu.Lock()
	g.sessions[ctx.Event.GroupID] = true
	g.mu.Unlock()

	diceRecv, diceCancel := zero.NewFutureEvent("message", 999, false,
		zero.OnlyGroup,
		zero.RegexRule(`^\d{1,`+strconv.Itoa(len(duel))+`}$|开始投币`),
		zero.CheckGroup(ctx.Event.GroupID),
		zero.CheckUser(uidlist...)).Repeat()
	defer diceCancel()

	ctx.SendChain(message.Text("游戏开始,请参游人员宣言正面硬币数量或开房人发送\"开始投币\"开始投币"))

	timeout = time.NewTimer(iquartersTimeout)
	defer timeout.Stop()

	mun := 0
	answer := ""

	for {
		select {
		case <-timeout.C:
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("等待超时,游戏取消\n咕之人扣除 6 星之尘")))
			for uid, guess := range duel {
				if guess == -1 {
					err := wallet.InsertWalletOf(uid, -6)
					if err != nil {
						ctx.SendChain(message.At(uid), message.Text("ERROR]:", err))
					}
				}
			}
			return
		case c := <-diceRecv:
			eventID := c.Event.UserID
			answer = c.Event.Message.String()

			if answer != "开始投币" && duel[eventID] == -1 {
				mun++
				guess, _ := strconv.Atoi(answer)
				duel[eventID] = guess
				ctx.SendChain(message.Text("已记录你宣言的数目:", guess))
			}

			if answer == "开始投币" {
				if mun >= len(duel) {
					break
				}
				ctx.SendChain(message.Text("还有人没有宣言数量喔"))
			}
		}
	}

	g.throwCoinsAndSettle(ctx, duel)
}

func (g *iquartersGame) handleJoinGame(ctx *zero.Ctx) {
}

func (g *iquartersGame) handleStartGame(ctx *zero.Ctx) {
}

func (g *iquartersGame) throwCoinsAndSettle(ctx *zero.Ctx, duel map[int64]int) {
	positive := 0
	result := "\n"

	for i := 0; i < len(duel); i++ {
		if rand.Intn(2) == 1 {
			result += "正 "
			positive++
		} else {
			result += "反 "
		}
	}

	ctx.SendChain(message.Text("OK,我要开始投掷银币了～"))
	process.SleepAbout1sTo2s()
	ctx.SendChain(message.Text("一共投掷了", len(duel), "枚银币,其中正面的有", positive, "枚正面。\n具体结果如下", result))

	winmsg := message.Message{}
	othermsg := message.Message{}
	losemsg := message.Message{}

	for uid, guess := range duel {
		switch {
		case guess == positive:
			err := wallet.InsertWalletOf(uid, positive*iquartersMultiplier1)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("ERROR]:", err))
			}
			winmsg = append(winmsg, message.At(uid))
		case math.Abs(guess-positive) <= 2:
			err := wallet.InsertWalletOf(uid, positive*iquartersMultiplier2)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("ERROR]:", err))
			}
			othermsg = append(othermsg, message.At(uid))
		default:
			err := wallet.InsertWalletOf(uid, -iquartersPenalty)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("ERROR]:", err))
			}
			losemsg = append(losemsg, message.At(uid))
		}
	}

	msg := message.Message{}
	if len(winmsg) != 0 {
		msg = append(msg, winmsg...)
		msg = append(msg, message.Text(fmt.Sprintf("\n恭喜获得 %d 星之尘\n\n", positive*iquartersMultiplier1)))
	}
	if len(othermsg) != 0 {
		msg = append(msg, othermsg...)
		msg = append(msg, message.Text(fmt.Sprintf("\n恭喜获得 %d 星之尘\n\n", positive*iquartersMultiplier2)))
	}
	if len(losemsg) != 0 {
		msg = append(msg, losemsg...)
		msg = append(msg, message.Text(fmt.Sprintf("\n很遗憾失去了 %d 星之尘\n\n", iquartersPenalty)))
	}

	ctx.Send(msg)
}

func (g *iquartersGame) endGame(groupID int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.sessions, groupID)
}
