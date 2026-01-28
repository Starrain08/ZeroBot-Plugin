package partygame

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	dieMsg = []string{
		"很不幸, 你死了......",
		"砰...很不幸, 你死了......",
		"你死了...",
		"很不幸, 你死了......",
		"你扣下了扳机\n你死了...",
		"你拿着手枪掂了掂, 你赌枪里没有子弹\n然后很不幸, 你死了...",
		"你是一个有故事的人, 但是子弹并不想知道这些, 它只看见了白花花的脑浆\n你死了",
		"你没有想太多, 扣下了扳机。你感觉到有什么东西从你的旁边飞过, 然后意识陷入了黑暗\n你死了",
		"大多数人对自己活着并不心存感激, 但你不再是了\n你死了...",
		"你举起了枪又放下, 然后又举了起来, 你的内心在挣扎, 但是你还是扣下了扳机, 你死了...",
		"你开枪之前先去吃了杯泡面\n然后很不幸, 你死了...",
		"你对此胸有成竹, 你曾经在精神病院向一个老汉学习了用手指夹住子弹的功夫\n然后很不幸你没夹住手滑了, 死了...",
		"今天的风儿很喧嚣, 沙尘能让眼睛感到不适。你去揉眼睛的时候手枪走火, 贯穿了你的小腹。然后很不幸, 你死了...",
		"我会死吗?我死了吗?你正这样想着\n然后很不幸, 你死了...",
		"漆黑的眩晕中, 心脏渐渐窒息无力, 彻骨的寒冷将你包围\n很不幸, 你死了...",
	}

	aliveMsg = []string{
		"你活了下来, 下一位",
		"你扣动扳机, 无事发生\n你活了下来",
		"你自信的扣动了扳机, 正如你所想象的那样\n你活了下来, 下一位",
		"你感觉命运女神在向你招手\n然后, 你活了下来, 下一位",
		"你吃了杯泡面发现没有调料, 你觉得不幸的你恐怕是死定了\n然后, 你活了下来, 下一位",
		"人和人的体质不能一概而论, 你在极度愤怒下, 扣下了扳机。利用扳机扣下和触发子弹的时间差, 手指一个加速硬生生扣断了它。\n然后, 你活了下来, 下一位",
		"你曾经在精神病院向一个老汉学习了用手指夹住子弹的功夫\n然后, 子弹并没有射出, 你活了下来, 下一位",
		"你曾经在精神病院向一个老汉学习过用手指夹住射出子弹的功夫, 在子弹射出的一瞬间, 你把他塞了回去\n你活了下来, 下一位",
	}

	rouletteHelpText = "- 创建轮盘赌\n- 加入轮盘赌\n- 开始轮盘赌\n- 开火\n- 终止轮盘赌"
)

func init() {
	engine := control.Register("roulette", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "轮盘赌",
		Help:              rouletteHelpText,
		PrivateDataFolder: "roulette",
	})

	filePath := engine.DataFolder() + "rate.json"
	_ = os.Remove(filePath)

	if err := initializeDataPath(filePath); err != nil {
		logrus.Errorf("[Roulette]初始化数据文件失败: %v", err)
		return
	}

	logrus.Infof("[Roulette]轮盘赌插件初始化成功，数据文件路径: %s", filePath)

	checkSession := func(ctx *zero.Ctx) bool {
		session, err := findSessionByGroupID(filePath, ctx.Event.GroupID)
		if err != nil {
			return true
		}

		switch ctx.Event.RawMessage {
		case "创建轮盘赌":
			if session.GroupID == 0 {
				return true
			}
			if session.IsValid {
				if session.IsExpired() {
					_ = session.Terminate()
					return true
				}
				sendErrorMessage(ctx, "轮盘赌游戏已经开始了")
				return false
			}
		default:
			if session.GroupID != ctx.Event.GroupID {
				return false
			}
			if session.IsValid {
				if session.IsExpired() {
					sendErrorMessage(ctx, "轮盘赌游戏已过期, 请重新开始")
					_ = session.Terminate()
					return false
				}
				sendErrorMessage(ctx, "轮盘赌游戏已经开始了")
				return false
			}
		}
		return true
	}

	engine.OnFullMatch(`创建轮盘赌`, zero.OnlyGroup, checkSession).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			groupID := ctx.Event.GroupID
			userID := ctx.Event.UserID

			// 验证输入参数
			if err := ValidateOnly(groupID); err != nil {
				sendErrorMessage(ctx, "无效的群组ID: "+err.Error())
				return
			}

			if err := ValidateOnly(userID); err != nil {
				sendErrorMessage(ctx, "无效的用户ID: "+err.Error())
				return
			}

			if err := createNewSession(filePath, groupID, userID); err != nil {
				sendErrorMessage(ctx, "创建游戏失败: "+err.Error())
				return
			}

			totalUsers := 1
			maxUsers := int(MaxPlayers)
			canJoin := maxUsers - totalUsers
			sendSuccessMessage(ctx, fmt.Sprintf("游戏开始, 目前有%d位玩家, 最多还能再加入%d名玩家, 发送\"加入轮盘赌\"加入游戏", totalUsers, canJoin))
		})

	engine.OnFullMatch("加入轮盘赌", zero.OnlyGroup, checkSession).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			defer recoverWithError(ctx, "加入轮盘赌")

			groupID := ctx.Event.GroupID
			userID := ctx.Event.UserID

			safeCall(ctx, func() error {
				session, err := findSessionByGroupID(filePath, groupID)
				if err != nil {
					return fmt.Errorf("获取游戏信息失败: %w", err)
				}

				if err := validatePlayerJoin(session, userID); err != nil {
					return err
				}

				if err := session.AddPlayer(userID); err != nil {
					return fmt.Errorf("加入游戏失败: %w", err)
				}

				newCount := session.GetUserCount()
				logrus.Infof("[Roulette]用户 %d 加入游戏，当前玩家数量: %d", userID, newCount)
				sendSuccessMessage(ctx, "成功加入,目前已有"+strconv.Itoa(newCount)+"位玩家,发送\"开始轮盘赌\"进行游戏")
				return nil
			})
		})

	engine.OnFullMatch("开始轮盘赌", zero.OnlyGroup, checkSession).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			defer recoverWithError(ctx, "开始轮盘赌")

			groupID := ctx.Event.GroupID
			userID := ctx.Event.UserID

			safeCall(ctx, func() error {
				session, err := findSessionByGroupID(filePath, groupID)
				if err != nil {
					return fmt.Errorf("获取游戏信息失败: %w", err)
				}

				if !session.IsPlayerInSession(userID) {
					return ErrPlayerNotFound
				}

				if err := validateSession(ctx, session); err != nil {
					return err
				}

				if err := session.ShufflePlayersOrder(); err != nil {
					return fmt.Errorf("初始化游戏失败: %w", err)
				}

				session.IsValid = true
				if err := saveSession(filePath, session); err != nil {
					return fmt.Errorf("启动游戏失败: %w", err)
				}

				sendSuccessMessage(ctx, "游戏开始,"+RouletteRule+"现在请"+strconv.FormatInt(session.Users[0], 10)+"开火")
				logrus.Infof("[Roulette]游戏开始，创建者: %d, 玩家数量: %d", userID, session.GetUserCount())
				return nil
			})

			gameCtx, cancel := context.WithTimeout(context.Background(), TimeoutDuration)
			defer cancel()

			warningTimer := time.NewTimer(WarningDuration)
			timeoutTimer := time.NewTimer(TimeoutDuration)

			defer warningTimer.Stop()
			defer timeoutTimer.Stop()

			stop, cancelStop := zero.NewFutureEvent("message", 8, true,
				zero.FullMatchRule("终止轮盘赌"),
				zero.AdminPermission).
				Repeat()
			defer cancelStop()

			next := zero.NewFutureEvent("message", 999, false, zero.FullMatchRule("开火"),
				zero.OnlyGroup, zero.CheckGroup(ctx.Event.GroupID))
			recv, cancel := next.Repeat()
			defer cancel()

			for {
				select {
				case <-warningTimer.C:
					sendErrorMessage(ctx, "轮盘赌, 还有15s过期")

				case <-timeoutTimer.C:
					if timeoutSession, err := findSessionByGroupID(filePath, groupID); err == nil {
						_ = timeoutSession.Terminate()
					}
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("轮盘赌超时, 游戏结束..."),
						),
					)
					return

				case <-stop:
					if stopSession, err := findSessionByGroupID(filePath, groupID); err == nil {
						_ = stopSession.Terminate()
					}
					sendSuccessMessage(ctx, "轮盘赌已终止")
					return

				case c := <-recv:
					warningTimer.Reset(WarningDuration)
					timeoutTimer.Reset(TimeoutDuration)

					currentSession, err := findSessionByGroupID(filePath, groupID)
					if err != nil {
						continue
					}

					uid := c.Event.UserID
					if !currentSession.IsPlayerInSession(uid) {
						sendErrorMessage(ctx, "你未参与游戏")
						continue
					}

					if !currentSession.IsPlayerTurn(uid) {
						sendErrorMessage(ctx, "未轮到你开火")
						continue
					}

					if currentSession.GetRemainingCartridges() == 1 {
						winnerID := uid
						if len(currentSession.Users) > 1 {
							winnerID = currentSession.Users[1]
						}
						_ = currentSession.Terminate()
						ctx.SendChain(message.At(winnerID), message.Text("长舒了一口气, 并反手击毙了"), message.At(uid))
						c.Event.UserID = uid
						go getTruthOrDare(c)
						return
					}

					isDead, err := currentSession.Fire()
					if err != nil {
						sendErrorMessage(ctx, "开火失败")
						continue
					}

					if isDead {
						_ = currentSession.Terminate()
						sendSuccessMessage(ctx, randomChoice(dieMsg))
						go getTruthOrDare(c)
						return
					}

					aliveText := randomChoice(aliveMsg)
					nextUser := currentSession.Users[1]
					sendSuccessMessage(ctx, aliveText+",轮到"+strconv.FormatInt(nextUser, 10)+"开火")
				case <-gameCtx.Done():
					return
				}
			}
		})
}
