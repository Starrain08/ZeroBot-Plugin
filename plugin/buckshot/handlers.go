package buckshot

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func handleCreateGame(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	if game, exists := getGame(gid); !exists {
		newGame := &Game{
			player1: &Player{
				name: string(ctx.CardOrNickName(ctx.Event.UserID)),
				id:   ctx.Event.UserID,
				hp:   6,
			},
			status: "waiting",
		}
		setGame(gid, newGame)
		ctx.SendChain(message.Text("══恶魔轮盘══\n游戏创建成功\n玩家1：", ctx.CardOrNickName(ctx.Event.UserID), "(", ctx.Event.UserID, ")\n玩家2：等待中\n发送\"恶魔轮盘.加入游戏\"以加入游戏"))
	} else if game.status == "waiting" {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道已有游戏正在等待玩家\n发送\"恶魔轮盘.加入游戏\"以加入游戏"))
	} else {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道已有游戏正在进行中"))
	}
}

func handleJoinGame(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	game, exists := getGame(gid)
	if !exists {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道没有可以加入的游戏\n发送\"恶魔轮盘.创建游戏\"以创建游戏"))
		return
	}
	if game.status != "waiting" {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道已有游戏正在进行中"))
		return
	}
	if game.player1.id == ctx.Event.UserID {
		ctx.SendChain(message.Text("══恶魔轮盘══\n你不能加入你自己创建的游戏"))
		return
	}
	if cancel, exists := dontDispose[gid]; exists {
		cancel()
		delete(dontDispose, gid)
	}
	game.mu.Lock()
	defer game.mu.Unlock()
	game.player2 = &Player{
		name: string(ctx.CardOrNickName(ctx.Event.UserID)),
		id:   ctx.Event.UserID,
		hp:   6,
	}
	game.status = "full"
	ctx.SendChain(message.Text("══恶魔轮盘══\n游戏开始\n玩家1：", game.player1.name, "(", game.player1.id, ")\n玩家2：", ctx.CardOrNickName(ctx.Event.UserID), "(", ctx.Event.UserID, ")\n由玩家1@", game.player1.id, "发送\"恶魔轮盘.开始游戏\"以开始游戏"))
}

func handleStartGame(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	game, exists := getGame(gid)
	if !exists {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道没有可以开始的游戏\n发送\"恶魔轮盘.创建游戏\"以创建游戏"))
		return
	}
	if game.player1.id != ctx.Event.UserID {
		ctx.SendChain(message.Text("══恶魔轮盘══\n只有玩家1可以开始游戏"))
		return
	}
	if game.status != "full" {
		ctx.SendChain(message.Text("══恶魔轮盘══\n正在等待玩家2\n发送\"恶魔轮盘.加入游戏\"以加入游戏"))
		return
	}
	game.mu.Lock()
	defer game.mu.Unlock()
	game.status = "started"
	game.bullet = shuffle(bulletPatterns[rand.Intn(len(bulletPatterns))])
	game.currentTurn = rand.Intn(2) + 1
	game.double = false
	game.round = 0
	game.usedHandcuff = false
	itemCount := rand.Intn(3) + 3
	for i := 0; i < itemCount-1; i++ {
		game.getPlayer(game.currentTurn).items = append(game.getPlayer(game.currentTurn).items, getRandomItem(game.round))
	}
	for i := 0; i < itemCount; i++ {
		game.getPlayer(3 - game.currentTurn).items = append(game.getPlayer(3-game.currentTurn).items, getRandomItem(game.round))
	}
	firstPlayer := game.getPlayer(game.currentTurn)
	ctx.SendChain(message.Text("══恶魔轮盘══\n游戏开始\n玩家1：@", game.player1.id, "\n玩家2：@", game.player2.id, "\n@", firstPlayer.id, "先手\n先手方获得", strconv.Itoa(itemCount-1), "个道具，后手方获得", strconv.Itoa(itemCount), "个道具\n枪内目前有", strconv.Itoa(countBullets(game.bullet, "实弹")), "发实弹和", strconv.Itoa(countBullets(game.bullet, "空包弹")), "发空包弹\n发送\"恶魔轮盘.对战信息\"以查看当前对战的游戏信息（如血量，道具等）"))
}

func handleBattleInfo(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	game, exists := getGame(gid)
	if !exists || game.status != "started" {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道没有正在进行的游戏\n发送\"恶魔轮盘.创建游戏\"以创建游戏"))
		return
	}
	game.mu.Lock()
	defer game.mu.Unlock()
	var result strings.Builder
	result.WriteString("══恶魔轮盘══\n--血量--\n")
	result.WriteString("玩家1(" + game.player1.name + ")：" + strconv.Itoa(game.player1.hp) + "/6点\n")
	result.WriteString("玩家2(" + game.player2.name + ")：" + strconv.Itoa(game.player2.hp) + "/6点\n\n")
	result.WriteString("--玩家1的道具 (" + strconv.Itoa(len(game.player1.items)) + "/8)--\n")
	if alwaysShowDesc {
		for _, item := range game.player1.items {
			if itemDesc, exists := itemList[item]; exists {
				result.WriteString(item + "(" + itemDesc.description + ")\n")
			}
		}
		result.WriteString("\n--玩家2的道具 (" + strconv.Itoa(len(game.player2.items)) + "/8)--\n")
		for _, item := range game.player2.items {
			if itemDesc, exists := itemList[item]; exists {
				result.WriteString(item + "(" + itemDesc.description + ")\n")
			}
		}
	} else {
		if len(game.player1.items) > 0 {
			result.WriteString(strings.Join(game.player1.items, ", "))
		} else {
			result.WriteString("无")
		}
		result.WriteString("\n\n--玩家2的道具 (" + strconv.Itoa(len(game.player2.items)) + "/8)--\n")
		if len(game.player2.items) > 0 {
			result.WriteString(strings.Join(game.player2.items, ", "))
		} else {
			result.WriteString("无")
		}
		result.WriteString("\n\n发送\"恶魔轮盘.道具说明 [道具名]\"以查看道具描述")
	}
	result.WriteString("\n发送道具名以使用道具\n发送\"自己\"或\"对方\"以选择向谁开枪")
	ctx.SendChain(message.Text(result.String()))
}

func handleItemDescription(ctx *zero.Ctx) {
	matched := ctx.State["regex_matched"].([]string)
	if len(matched) < 2 {
		ctx.SendChain(message.Text("道具不存在"))
		return
	}
	itemName := matched[1]
	item, exists := itemList[itemName]
	if !exists {
		ctx.SendChain(message.Text("道具不存在"))
		return
	}
	ctx.SendChain(message.Text(item.description2))
}

func handleEndGame(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	game, exists := getGame(gid)
	if !exists {
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前频道没有已创建或正在进行的游戏"))
		return
	}
	game.mu.Lock()
	defer game.mu.Unlock()
	if !isAdmin(ctx) && game.player1.id != ctx.Event.UserID && (game.player2 == nil || game.player2.id != ctx.Event.UserID) {
		ctx.SendChain(message.Text("══恶魔轮盘══\n只有当前游戏中的玩家或管理员才能结束游戏"))
		return
	}
	deleteGame(gid)
	if cancel, exists := dontDispose[gid]; exists {
		cancel()
		delete(dontDispose, gid)
	}
	ctx.SendChain(message.Text("══恶魔轮盘══\n游戏已被@", ctx.Event.UserID, "结束"))
}

func handleShoot(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	game, exists := getGame(gid)
	if !exists || game.status != "started" {
		return
	}
	if !isCurrentPlayer(game, ctx) {
		return
	}
	matched := ctx.State["regex_matched"].([]string)
	if len(matched) < 2 {
		return
	}
	target := matched[1]
	game.mu.Lock()
	defer game.mu.Unlock()
	if len(game.bullet) == 0 {
		ctx.SendChain(message.Text("枪内没有子弹"))
		return
	}
	bullet := game.bullet[len(game.bullet)-1]
	game.bullet = game.bullet[:len(game.bullet)-1]
	result := "══恶魔轮盘══\n你将枪口对准了" + target + "\n扣下扳机，是" + bullet + "\n"
	dead := false
	if bullet == "实弹" {
		damage := 1
		if game.double {
			damage = 2
		}
		if target == "自己" {
			player := game.getPlayer(game.currentTurn)
			player.hp -= damage
			result += "你损失了" + strconv.Itoa(damage) + "点生命值"
			if player.hp <= 0 {
				dead = true
				deleteGame(gid)
				muteConfig := getMuteConfig()
				if muteConfig.enabled && ctx.Event.GroupID != 0 {
					ctx.SetGroupBan(ctx.Event.GroupID, player.id, int64(muteConfig.duration*60))
				}
				ctx.SendChain(message.Text(result), message.Text("\n══恶魔轮盘══\n@", player.id, "倒在了桌前\n@", game.getPlayer(3-game.currentTurn).id, "获得了胜利\n游戏结束"))
				return
			}
		} else {
			other := game.getPlayer(3 - game.currentTurn)
			other.hp -= damage
			result += "对方损失了" + strconv.Itoa(damage) + "点生命值"
			if other.hp <= 0 {
				dead = true
				deleteGame(gid)
				muteConfig := getMuteConfig()
				if muteConfig.enabled && ctx.Event.GroupID != 0 {
					ctx.SetGroupBan(ctx.Event.GroupID, other.id, int64(muteConfig.duration*60))
				}
				ctx.SendChain(message.Text(result), message.Text("\n══恶魔轮盘══\n@", other.id, "倒在了桌前\n@", game.getPlayer(game.currentTurn).id, "获得了胜利\n游戏结束"))
				return
			}
		}
	}
	if !dead {
		if bullet == "空包弹" && target == "自己" {
			result += "接下来还是你的回合"
		} else {
			other := game.getPlayer(3 - game.currentTurn)
			if !other.handcuff {
				game.currentTurn = 3 - game.currentTurn
				player := game.getPlayer(game.currentTurn)
				result += "\n接下来是@" + strconv.FormatInt(player.id, 10) + "的回合"
				game.usedHandcuff = false
			} else {
				other.handcuff = false
				result += "\n因为对方被手铐铐住了，接下来还是你的回合"
			}
		}
		ctx.SendChain(message.Text(result))
		if len(game.bullet) == 0 {
			roundResult := nextRound(game, itemList)
			for _, r := range roundResult.result {
				ctx.SendChain(message.Text(r))
			}
		}
		game.double = false
	}
}

func handleUseItem(ctx *zero.Ctx) {
	gid := getChannelIDFromEvent(ctx.Event.GroupID, ctx.Event.UserID)
	game, exists := getGame(gid)
	if !exists || game.status != "started" {
		return
	}

	msg := getMessageContent(ctx)
	if msg == "" {
		return
	}

	game.mu.Lock()

	if game.waitingForItem {
		waitingPlayer := game.getPlayer(game.waitingPlayer)
		if ctx.Event.UserID == waitingPlayer.id {
			otherPlayer := game.getPlayer(3 - game.waitingPlayer)
			if timeoutChan, exists := epinephrineTimeouts[gid]; exists {
				select {
				case timeoutChan <- struct{}{}:
					delete(epinephrineTimeouts, gid)
				default:
				}
			}

			hasItem := false
			for _, item := range otherPlayer.items {
				if item == msg {
					hasItem = true
					break
				}
			}
			if !hasItem {
				ctx.SendChain(message.Text("对方没有这个道具，已取消使用"))
				game.waitingForItem = false
				game.mu.Unlock()
				return
			}
			if msg == "肾上腺素" {
				ctx.SendChain(message.Text("不能选择肾上腺素"))
				game.waitingForItem = false
				game.mu.Unlock()
				return
			}

			game.waitingForItem = false

			if item, exists := itemList[msg]; exists {
				success, result := item.use(game, game.waitingPlayer, itemList)
				if success {
					otherPlayer.removeItem(msg)
				}
				game.mu.Unlock()
				for _, r := range result {
					ctx.SendChain(message.Text(r))
				}
				if otherPlayer.hp <= 0 {
					deleteGame(gid)
					muteConfig := getMuteConfig()
					if muteConfig.enabled && ctx.Event.GroupID != 0 {
						ctx.SetGroupBan(ctx.Event.GroupID, otherPlayer.id, int64(muteConfig.duration*60))
					}
					ctx.SendChain(message.Text("\n══恶魔轮盘══\n@", otherPlayer.id, "倒在了桌前\n@", waitingPlayer.id, "获得了胜利\n游戏结束"))
				}
				return
			}
			game.mu.Unlock()
			return
		}
		game.mu.Unlock()
		return
	}

	game.mu.Unlock()

	if !isCurrentPlayer(game, ctx) {
		return
	}

	game.mu.Lock()
	player := game.getPlayer(game.currentTurn)
	hasItem := false
	for _, item := range player.items {
		if item == msg {
			hasItem = true
			break
		}
	}
	if !hasItem {
		game.mu.Unlock()
		return
	}

	if msg == "肾上腺素" {
		player.removeItem(msg)
		game.waitingForItem = true
		game.waitingPlayer = game.currentTurn
		initialTurn := game.currentTurn
		game.mu.Unlock()

		ctx.SendChain(message.Text("你给自己来了一针肾上腺素，请在30秒内发送你想选择的对方道具名（不能选择肾上腺素）"))

		timeoutChan := make(chan struct{})
		epinephrineTimeouts[gid] = timeoutChan

		go func() {
			select {
			case <-time.After(30 * time.Second):
				gamesMutex.RLock()
				currentGame, exists := games[gid]
				gamesMutex.RUnlock()
				if exists {
					currentGame.mu.Lock()
					if currentGame.waitingForItem && currentGame.waitingPlayer == initialTurn {
						currentGame.waitingForItem = false
						currentGame.mu.Unlock()
						ctx.SendChain(message.Text("选择超时，已取消使用"))
					} else {
						currentGame.mu.Unlock()
					}
				}
				delete(epinephrineTimeouts, gid)
			case <-timeoutChan:
			}
		}()

		return
	}

	if item, exists := itemList[msg]; exists {
		success, result := item.use(game, game.currentTurn, itemList)
		if success {
			player.removeItem(msg)
		}
		game.mu.Unlock()
		for _, r := range result {
			ctx.SendChain(message.Text(r))
		}
		if player.hp <= 0 {
			deleteGame(gid)
			muteConfig := getMuteConfig()
			if muteConfig.enabled && ctx.Event.GroupID != 0 {
				ctx.SetGroupBan(ctx.Event.GroupID, player.id, int64(muteConfig.duration*60))
			}
			ctx.SendChain(message.Text("\n══恶魔轮盘══\n@", player.id, "倒在了桌前\n@", game.getPlayer(3-game.currentTurn).id, "获得了胜利\n游戏结束"))
		}
		return
	}
	game.mu.Unlock()
}

func isCurrentPlayer(game *Game, ctx *zero.Ctx) bool {
	game.mu.Lock()
	defer game.mu.Unlock()
	player := game.getPlayer(game.currentTurn)
	return player.id == ctx.Event.UserID
}

func getMessageContent(ctx *zero.Ctx) string {
	if args, ok := ctx.State["args"]; ok {
		if str, ok := args.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	msgStr := ctx.Event.Message.String()
	msgStr = strings.ReplaceAll(msgStr, "[CQ:at,qq=", "")
	msgStr = strings.ReplaceAll(msgStr, "]", "")
	return strings.TrimSpace(msgStr)
}

func handleSetMuteConfig(ctx *zero.Ctx) {
	if !isAdmin(ctx) {
		ctx.SendChain(message.Text("══恶魔轮盘══\n只有管理员才能配置禁言设置"))
		return
	}

	msgStr := ctx.Event.Message.String()

	if strings.Contains(msgStr, "恶魔轮盘.设置禁言") {
		if strings.Contains(msgStr, "开启") {
			duration := 60
			parts := strings.Split(msgStr, "开启")
			if len(parts) > 1 {
				args := strings.Fields(parts[1])
				for _, arg := range args {
					num, err := strconv.Atoi(arg)
					if err == nil && num > 0 {
						duration = num
						break
					}
				}
			}
			setMuteConfig(true, duration)
			ctx.SendChain(message.Text("══恶魔轮盘══\n禁言惩罚已开启，输家将被禁言", strconv.Itoa(duration), "秒"))
		} else if strings.Contains(msgStr, "关闭") {
			setMuteConfig(false, 0)
			ctx.SendChain(message.Text("══恶魔轮盘══\n禁言惩罚已关闭"))
		}
	} else if strings.Contains(msgStr, "恶魔轮盘.禁言设置") {
		config := getMuteConfig()
		status := "关闭"
		if config.enabled {
			status = "开启"
		}
		ctx.SendChain(message.Text("══恶魔轮盘══\n当前禁言设置：", status, "\n禁言时长：", strconv.Itoa(config.duration), "秒\n\n使用\"恶魔轮盘.设置禁言开启 [时长]\"开启禁言\n使用\"恶魔轮盘.设置禁言关闭\"关闭禁言"))
	}
}
