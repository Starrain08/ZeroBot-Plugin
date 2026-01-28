package partygame

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type tdata struct {
	Version string `json:"version"`
	Data    []struct {
		Name string   `json:"name"`
		Des  string   `json:"des"`
		Tags []string `json:"tags"`
	} `json:"data"`
}

var (
	action              tdata
	question            tdata
	punishMap           = make(map[string]string)
	punishMutex         sync.RWMutex
	truthOrDareHelpText = "真心话大冒险\n- 来点乐子[@xxx]\n- 饶恕[@xxx]\n- 惩罚[@xxx]\n- 反省[@xxx]"
)

func init() {
	engine := control.Register("truthordare", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             truthOrDareHelpText,
		PublicDataFolder: "Truthordare",
	})

	actionData, err := engine.GetLazyData("action.json", false)
	if err != nil {
		logrus.Errorf("[Truthordare]加载action.json失败: %v", err)
		return
	}

	questionData, err := engine.GetLazyData("question.json", false)
	if err != nil {
		logrus.Errorf("[Truthordare]加载question.json失败: %v", err)
		return
	}

	if err := json.Unmarshal(actionData, &action); err != nil {
		logrus.Errorf("[Truthordare]解析action.json失败: %v", err)
		return
	}

	if err := json.Unmarshal(questionData, &question); err != nil {
		logrus.Errorf("[Truthordare]解析question.json失败: %v", err)
		return
	}

	logrus.Infof("[Truthordare]加载 %d 条真心话, %d 条大冒险", len(question.Data), len(action.Data))
	logrus.Infof("[Truthordare]真心话大冒险插件初始化成功")

	engine.OnRegex(`^(真心话大冒险|来点刺激|来点乐子)`).Handle(func(ctx *zero.Ctx) {
		defer recoverWithError(ctx, "真心话大冒险触发")

		// 验证输入
		if err := validateStringInput(ctx.Event.Message.ExtractPlainText(), 3, 20, "触发命令"); err != nil {
			sendErrorMessage(ctx, "无效的命令格式")
			return
		}

		targetUserID := parseTargetUserID(ctx)

		if err := ValidateOnly(targetUserID); err != nil {
			sendErrorMessage(ctx, "无效的目标用户ID")
			return
		}

		key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, targetUserID)

		punishMutex.RLock()
		if v, exists := punishMap[key]; exists {
			punishMutex.RUnlock()
			sendSuccessMessageWithAt(ctx, targetUserID, "罪行尚未被饶恕, 赎罪方式是"+v)
			return
		}
		punishMutex.RUnlock()

		ctx.Event.UserID = targetUserID
		logrus.Infof("[TruthOrDare]用户 %d 触发真心话大冒险", targetUserID)
		getTruthOrDare(ctx)
	})

	engine.OnRegex(`^(饶恕|阿门|释放|原谅|赦免)`, zero.AdminPermission, zero.OnlyGroup).Handle(func(ctx *zero.Ctx) {
		targetUserID := parseTargetUserID(ctx)
		key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, targetUserID)

		punishMutex.Lock()
		delete(punishMap, key)
		punishMutex.Unlock()
		sendSuccessMessageWithAt(ctx, targetUserID, "恭喜你恢复自由之身")
	})

	engine.OnRegex(`^(惩罚|降下神罚)`, zero.AdminPermission, zero.OnlyGroup).Handle(func(ctx *zero.Ctx) {
		targetUserID := parseTargetUserID(ctx)
		ctx.Event.UserID = targetUserID
		getTruthOrDare(ctx)
	})

	engine.OnRegex(`^(反省|检查罪行)`, zero.OnlyGroup).Handle(func(ctx *zero.Ctx) {
		targetUserID := parseTargetUserID(ctx)
		key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, targetUserID)

		punishMutex.RLock()
		if v, exists := punishMap[key]; exists {
			sendSuccessMessageWithAt(ctx, targetUserID, "你是罪人, 赎罪方式是"+v)
		}
		punishMutex.RUnlock()

		role := ctx.GetGroupMemberInfo(ctx.Event.GroupID, targetUserID, true).Get("role").String()
		ctx.Event.UserID = targetUserID

		if zero.SuperUserPermission(ctx) || role != "member" {
			sendSuccessMessageWithAt(ctx, targetUserID, "你是上帝")
		}

		if !zero.SuperUserPermission(ctx) && role == "member" {
			punishMutex.RLock()
			if _, exists := punishMap[key]; !exists {
				sendSuccessMessageWithAt(ctx, targetUserID, "你是平民")
			}
			punishMutex.RUnlock()
		}
	})
}

func getAction() string {
	if len(action.Data) == 0 {
		return "大冒险(暂无数据)"
	}
	return randomChoice(action.Data).Name
}

func getQuestion() string {
	if len(question.Data) == 0 {
		return "真心话(暂无数据)"
	}
	return randomChoice(question.Data).Name
}

func getActionOrQuestion() string {
	if len(action.Data) == 0 && len(question.Data) == 0 {
		return "真心话大冒险(暂无数据)"
	}
	if len(action.Data) == 0 {
		return getQuestion()
	}
	if len(question.Data) == 0 {
		return getAction()
	}
	if randSource.Intn(2) == 0 {
		return getAction()
	}
	return getQuestion()
}

func getTruthOrDare(ctx *zero.Ctx) {
	defer recoverWithError(ctx, "getTruthOrDare")

	next, cancel := zero.NewFutureEvent("message", 999, false, ctx.CheckSession(), zero.FullMatchRule("真心话", "大冒险")).Repeat()
	defer cancel()

	key := fmt.Sprintf("%v-%v", ctx.Event.GroupID, ctx.Event.UserID)
	sendSuccessMessageWithAt(ctx, ctx.Event.UserID, "你将受到严峻的惩罚, 请选择惩罚, 真心话还是大冒险?")

	timeout := time.After(ResponseTimeout)

	for {
		select {
		case <-timeout:
			botName := "ZeroBot"
			if len(zero.BotConfig.NickName) > 0 {
				botName = zero.BotConfig.NickName[0]
			}
			sendSuccessMessage(ctx, "时间太久啦！"+botName+"帮你选择")
			punishment := getActionOrQuestion()

			punishMutex.Lock()
			punishMap[key] = punishment
			punishMutex.Unlock()

			sendSuccessMessageWithAt(ctx, ctx.Event.UserID, "恭喜你获得\""+punishment+"\"的惩罚")
			logrus.Infof("[TruthOrDare]用户 %d 自动选择惩罚: %s", ctx.Event.UserID, punishment)
			return

		case c := <-next:
			msg := c.Event.Message.ExtractPlainText()

			if err := validateStringInput(msg, 2, 10, "惩罚类型"); err != nil {
				logrus.Warnf("[TruthOrDare]用户 %d 输入无效: %v", ctx.Event.UserID, err)
				continue
			}

			var punishment string

			switch msg {
			case "真心话":
				punishment = getQuestion()
			case "大冒险":
				punishment = getAction()
			default:
				continue
			}

			punishMutex.Lock()
			punishMap[key] = punishment
			punishMutex.Unlock()

			sendSuccessMessageWithAt(ctx, ctx.Event.UserID, "恭喜你获得\""+punishment+"\"的惩罚")
			logrus.Infof("[TruthOrDare]用户 %d 选择 %s: %s", ctx.Event.UserID, msg, punishment)
			return
		}
	}
}
