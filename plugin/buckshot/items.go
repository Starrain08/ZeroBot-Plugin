package buckshot

import (
	"math/rand"
	"strconv"
)

var (
	bulletPatterns = [][]string{
		{"实弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "空包弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "实弹", "空包弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "实弹", "空包弹", "空包弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "实弹", "实弹", "空包弹", "空包弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "实弹", "空包弹", "空包弹"},
		{"实弹", "实弹", "实弹", "实弹", "空包弹", "空包弹"},
	}
	itemList       map[string]Item
	alwaysShowDesc = true
)

type Item struct {
	description  string
	description2 string
	use          func(g *Game, player int, items map[string]Item) (bool, []string)
}

type RoundResult struct {
	game   *Game
	result []string
}

func getItemList() map[string]Item {
	return map[string]Item{
		"手锯": {
			description:  "下一发造成双倍伤害，不可叠加",
			description2: "下一发造成双倍伤害，不可叠加",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				g.double = true
				return true, []string{"你用手锯锯短了枪管，下一发将造成双倍伤害"}
			},
		},
		"放大镜": {
			description:  "查看当前膛内的子弹",
			description2: "查看当前膛内的子弹",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				if len(g.bullet) == 0 {
					return true, []string{"枪内没有子弹"}
				}
				bullet := g.bullet[len(g.bullet)-1]
				return true, []string{"你使用了放大镜，看到了膛内的子弹是" + bullet}
			},
		},
		"啤酒": {
			description:  "卸下当前膛内的子弹",
			description2: "卸下当前膛内的子弹",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				if len(g.bullet) == 0 {
					return true, []string{"枪内没有子弹"}
				}
				bullet := g.bullet[len(g.bullet)-1]
				g.bullet = g.bullet[:len(g.bullet)-1]
				if len(g.bullet) == 0 {
					roundResult := nextRound(g, items)
					return true, append([]string{"你喝下了啤酒，把当前膛内的子弹抛了出来，是一发" + bullet}, roundResult.result...)
				}
				return true, []string{"你喝下了啤酒，把当前膛内的子弹抛了出来，是一发" + bullet}
			},
		},
		"香烟": {
			description:  "恢复1点生命值",
			description2: "恢复1点生命值",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				p := g.getPlayer(player)
				if p.hp < 6 {
					p.hp++
					return true, []string{"你抽了一根香烟，恢复了1点生命值"}
				}
				return true, []string{"你抽了一根香烟，但什么都没有发生，因为你的生命值是满的"}
			},
		},
		"手铐": {
			description:  "跳过对方的下一回合",
			description2: "跳过对方的下一回合",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				if g.usedHandcuff {
					return false, []string{"一回合只能使用一次手铐"}
				}
				otherPlayer := g.getPlayer(3 - player)
				otherPlayer.handcuff = true
				g.usedHandcuff = true
				return true, []string{"你给对方上了手铐，对方的下一回合将被跳过"}
			},
		},
		"肾上腺素": {
			description:  "选择对方的1个道具并立刻使用，不能选择肾上腺素",
			description2: "选择对方的1个道具并立刻使用，不能选择肾上腺素",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				return true, []string{"你给自己来了一针肾上腺素，请在30秒内发送你想选择的对方道具名（不能选择肾上腺素）"}
			},
		},
		"过期药物": {
			description:  "50%概率恢复2点生命值，50%概率损失1点生命值",
			description2: "50%概率恢复2点生命值，50%概率损失1点生命值",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				p := g.getPlayer(player)
				if rand.Float64() < 0.5 {
					diff := 6 - p.hp
					if diff > 2 {
						diff = 2
					}
					p.hp += diff
					return true, []string{"你吃下了过期药物，感觉不错，恢复了2点生命值"}
				} else {
					p.hp--
					if p.hp <= 0 {
						return false, []string{"你吃下了过期药物，感觉不太对劲，但还没来得及思考就失去了意识"}
					}
					return true, []string{"你吃下了过期药物，感觉不太对劲，损失了1点生命值"}
				}
			},
		},
		"逆转器": {
			description:  "转换膛内的子弹，实弹变为空包弹，反之亦然",
			description2: "转换膛内的子弹，实弹变为空包弹，空包弹变为实弹",
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				if len(g.bullet) == 0 {
					return true, []string{"枪内没有子弹"}
				}
				lastBullet := g.bullet[len(g.bullet)-1]
				if lastBullet == "实弹" {
					g.bullet[len(g.bullet)-1] = "空包弹"
				} else {
					g.bullet[len(g.bullet)-1] = "实弹"
				}
				return true, []string{"你使用了逆转器，膛内的子弹发生了一些变化"}
			},
		},
		"骰子": {
			description: "掷一个六面骰子，根据点数触发不同的效果",
			description2: `掷一个六面骰子，根据点数触发以下效果
1：膛内子弹变为实弹
2：膛内子弹变为空包弹
3：随机触发某个道具的效果
4：恢复1滴血
5：损失1滴血
6：直接结束你的回合`,
			use: func(g *Game, player int, items map[string]Item) (bool, []string) {
				dice := rand.Intn(6) + 1
				switch dice {
				case 1:
					if len(g.bullet) == 0 {
						return true, []string{"枪内没有子弹"}
					}
					g.bullet[len(g.bullet)-1] = "实弹"
					return true, []string{"你骰出了1，膛内的子弹变成了实弹"}
				case 2:
					if len(g.bullet) == 0 {
						return true, []string{"枪内没有子弹"}
					}
					g.bullet[len(g.bullet)-1] = "空包弹"
					return true, []string{"你骰出了2，膛内的子弹变成了空包弹"}
				case 3:
					keys := []string{}
					for k := range items {
						if k != "骰子" {
							keys = append(keys, k)
						}
					}
					if len(keys) == 0 {
						return true, []string{"你骰出了3，但没有其他道具可以触发"}
					}
					item := keys[rand.Intn(len(keys))]
					success, result := items[item].use(g, player, items)
					return success, append([]string{"你骰出了3，转眼间骰子就变成了" + item}, result...)
				case 4:
					p := g.getPlayer(player)
					if p.hp < 6 {
						p.hp++
						return true, []string{"你骰出了4，这个数字让你感觉神清气爽，恢复了1点生命值"}
					}
					return true, []string{"你骰出了4，这个数字让你神清气爽，但什么都没有发生，因为你的生命值是满的"}
				case 5:
					p := g.getPlayer(player)
					p.hp--
					if p.hp <= 0 {
						return false, []string{"你骰出了5，你感觉这个数字不太行，但还没来得及思考就失去了意识"}
					}
					return true, []string{"你骰出了5，你感觉这个数字不太行，损失了1点生命值"}
				case 6:
					p := g.getPlayer(player)
					p.removeItem("骰子")
					g.currentTurn = 3 - g.currentTurn
					g.usedHandcuff = false
					g.double = false
					other := g.getPlayer(g.currentTurn)
					other.handcuff = false
					return false, []string{"你掷出了6，这个数字让你觉得被嘲讽了，急的你直接结束了回合\n接下来是@" + strconv.FormatInt(other.id, 10) + "的回合"}
				}
				return true, []string{}
			},
		},
	}
}

func getRandomItem(round int) string {
	keys := make([]string, 0, len(itemList))
	for k := range itemList {
		if round > 3 && (k == "香烟" || k == "过期药物") {
			continue
		}
		keys = append(keys, k)
	}
	if len(keys) == 0 {
		return ""
	}
	return keys[rand.Intn(len(keys))]
}

func nextRound(g *Game, items map[string]Item) RoundResult {
	g.round++
	var result []string
	if len(bulletPatterns) > 0 {
		g.bullet = shuffle(bulletPatterns[rand.Intn(len(bulletPatterns))])
	} else {
		g.bullet = []string{}
	}
	itemCount := rand.Intn(4) + 2
	for i := 0; i < itemCount; i++ {
		current := g.getPlayer(g.currentTurn)
		current.items = append(current.items, getRandomItem(g.round))
		other := g.getPlayer(3 - g.currentTurn)
		other.items = append(other.items, getRandomItem(g.round))
	}
	g.player1.items = g.player1.items[:min(8, len(g.player1.items))]
	g.player2.items = g.player2.items[:min(8, len(g.player2.items))]
	extraText := ""
	if g.round > 3 {
		extraText = "\n终极决战已开启，无法再获得回血道具"
	}
	result = append(result, "══恶魔轮盘══\n子弹打空了，进入下一轮"+extraText+"\n枪内目前有"+strconv.Itoa(countBullets(g.bullet, "实弹"))+"发实弹和"+strconv.Itoa(countBullets(g.bullet, "空包弹"))+"发空包弹\n双方获得"+strconv.Itoa(itemCount)+"个道具（道具上限为8个）")
	return RoundResult{game: g, result: result}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func shuffle(arr []string) []string {
	result := make([]string, len(arr))
	copy(result, arr)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

func countBullets(bullets []string, target string) int {
	count := 0
	for _, b := range bullets {
		if b == target {
			count++
		}
	}
	return count
}
