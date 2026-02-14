package dice

import (
	"math/rand/v2"
	"strconv"

	"github.com/FloatTech/floatbox/math"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	maxCheckTimes    = 10
	maxRollCount     = 100
	maxDiceSetting   = 1000
	defaultDiceValue = 20
)

var ruleHandlers = []func(r, num int64) string{
	rule0, rule1, rule2, rule3, rule4, rule5, rule6,
}

type intn interface {
	int | int8 | int16 | int32 | int64
}

func init() {
	engine.OnRegex(`^[。.][Rr][AaCc]\s*(([0-9]{1,2})#)?\s*(.*)\s+([0-9]{1,2})$`, zero.OnlyGroup).SetBlock(true).
		Handle(handleRandomCheck)
	engine.OnRegex(`^[。.]setcoc\s*([0-6]{1})`, zero.OnlyGroup).SetBlock(true).
		Handle(handleSetCOC)
	engine.OnRegex(`^[。.]set\s*([0-9]+)`, zero.OnlyGroup).SetBlock(true).
		Handle(handleSetDice)
	engine.OnRegex(`^[。.][Rr]\s*([0]*[1-9]+)?\s*[Dd]\s*([0]*[1-9]+)?`, zero.OnlyGroup).SetBlock(true).
		Handle(handleRoll)
}

func handleRandomCheck(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	nickname := ctx.CardOrNickName(ctx.Event.UserID)
	times := math.Str2Int64(matches[2])
	word := matches[3]
	num := math.Str2Int64(matches[4])

	if times == 0 {
		times = 1
	} else if times > maxCheckTimes {
		times = maxCheckTimes
	}

	var msg message.Message
	if times > 1 {
		msg = append(msg, message.Text("最多检定"+strconv.Itoa(maxCheckTimes)+"次哦~\n"))
	}
	msg = append(msg, message.Text(nickname, "进行", word, "检定:"))

	rule, _ := getGroupRule(ctx.Event.GroupID)
	for i := int64(0); i < times; i++ {
		rs := rand.Int64N(100) + 1
		win := evaluateRule(rs, num, rule)
		msg = append(msg, message.Text("\nD100=", rs, "/", num, " ", win))
	}
	ctx.SendChain(msg...)
}

func handleSetCOC(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	rule := math.Str2Int64(matches[1])

	r := &rsl{
		GrpID: ctx.Event.GroupID,
		Rule:  rule,
	}
	if err := db.Insert("rsl", r); err == nil {
		ctx.SendChain(message.Text("当前群聊房规设置为了", r.Rule))
	} else {
		ctx.SendChain(message.Text("出错啦: ", err))
	}
}

func handleSetDice(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	diceValue := math.Str2Int64(matches[1])

	if diceValue > maxDiceSetting {
		diceValue = maxDiceSetting
		ctx.SendChain(message.Text("最多" + strconv.Itoa(maxDiceSetting) + "哟~已自动设为" + strconv.Itoa(maxDiceSetting)))
	}

	d := &set{
		UserID: ctx.Event.UserID,
		D:      diceValue,
	}
	if err := db.Insert("set", d); err == nil {
		ctx.SendChain(message.Text("阁下默认骰子被设定为了", d.D))
	} else {
		ctx.SendChain(message.Text("出错啦: ", err))
	}
}

func handleRoll(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	r1 := math.Str2Int64(matches[1])
	d1 := math.Str2Int64(matches[2])

	if r1 == 0 {
		r1 = 1
	}
	if d1 == 0 {
		d1 = defaultDiceValue
		if userDice, err := getUserDiceSetting(ctx.Event.UserID); err == nil {
			d1 = userDice
		}
	}

	if r1 > maxRollCount {
		ctx.SendChain(message.Text("骰子太多啦~~数不过来了！"))
		return
	}

	var sum int64
	var res message.Message
	for i := int64(0); i < r1-1; i++ {
		roll := rand.Int64N(d1) + 1
		sum += roll
		res = append(res, message.Text("+", roll))
	}
	finalRoll := rand.Int64N(d1) + 1
	sum += finalRoll
	res = append(res, message.Text(finalRoll))
	ctx.SendChain(message.Text("阁下掷出了R", r1, "D", d1, "=", sum, "\n", res.String(), "=", sum))
}

func getGroupRule(groupID int64) (int64, error) {
	var r rsl
	err := db.Find("rsl", &r, "where gid = "+strconv.FormatInt(groupID, 10))
	if err != nil {
		return 0, err
	}
	return r.Rule, nil
}

func getUserDiceSetting(userID int64) (int64, error) {
	var d set
	err := db.Find("set", &d, "where uid = "+strconv.FormatInt(userID, 10))
	if err != nil {
		return 0, err
	}
	return d.D, nil
}

func evaluateRule(r, num, rule int64) string {
	ruleIdx := int(rule)
	if ruleIdx >= 0 && ruleIdx < len(ruleHandlers) {
		return ruleHandlers[ruleIdx](r, num)
	}
	return rule0(r, num)
}

func isDiceBigSuccess(r, num, rule int64) bool {
	switch rule {
	case 0:
		return r == 1
	case 1:
		return (num < 50 && r == 1) || (num >= 50 && r >= 1 && r <= 5)
	case 2:
		return r >= 1 && r <= 5 && r <= num
	case 3:
		return r >= 1 && r <= 5
	case 4:
		return r >= 1 && r <= 5 && r <= num/10
	case 5:
		return r >= 1 && r <= 2 && r <= num/5
	case 6:
		return (r == 1 && r <= num) || (r%11 == 0 && r <= num)
	default:
		return false
	}
}

func isDiceBigFailure(r, num, rule int64) bool {
	switch rule {
	case 0:
		return (num < 50 && r <= 100 && r >= 96) || (num >= 50 && r == 100)
	case 1:
		return (num < 50 && r < 100 && r > 96) || (num >= 50 && r == 100)
	case 2:
		return r >= 96 && r <= 100 && r > num
	case 3:
		return r >= 96 && r <= 100
	case 4:
		return (num < 50 && r >= 96+num/10) || (num >= 50 && r == 100)
	case 5:
		return (num < 50 && r >= 96 && r <= 100) || (num >= 50 && r >= 99 && r <= 100)
	case 6:
		return (r == 100 && r > num) || (r%11 == 0 && r > num)
	default:
		return false
	}
}

func isDiceVeryHardSuccess(r, num int64) bool {
	return r < num/5
}

func isDiceHardSuccess(r, num int64) bool {
	return r < num/2
}

func isDiceSuccess(r, num int64) bool {
	return r < num
}

func rules[T intn](r, num, rule T) string {
	ruleInt := int64(rule)
	rInt := int64(r)
	numInt := int64(num)
	return evaluateRule(rInt, numInt, ruleInt)
}

func rule0(r, num int64) string {
	return getDiceResult(r, num, 0)
}

func rule1(r, num int64) string {
	return getDiceResult(r, num, 1)
}

func rule2(r, num int64) string {
	return getDiceResult(r, num, 2)
}

func rule3(r, num int64) string {
	return getDiceResult(r, num, 3)
}

func rule4(r, num int64) string {
	return getDiceResult(r, num, 4)
}

func rule5(r, num int64) string {
	return getDiceResult(r, num, 5)
}

func rule6(r, num int64) string {
	return getDiceResult(r, num, 6)
}

func getDiceResult(r, num, rule int64) string {
	if isDiceBigSuccess(r, num, rule) {
		return "大成功"
	}
	if isDiceBigFailure(r, num, rule) {
		return "大失败"
	}
	if isDiceVeryHardSuccess(r, num) {
		return "极难成功"
	}
	if isDiceHardSuccess(r, num) {
		return "困难成功"
	}
	if isDiceSuccess(r, num) {
		return "成功"
	}
	return "失败"
}
