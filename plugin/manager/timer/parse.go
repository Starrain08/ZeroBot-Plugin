package timer

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

// GetTimerInfo 获得标准化定时字符串
func (t *Timer) GetTimerInfo() string {
	if t.Cron != "" {
		return fmt.Sprintf("[%d]%s", t.GrpID, t.Cron)
	}
	return fmt.Sprintf("[%d]%d月%d日%d周%d:%d", t.GrpID, t.Month(), t.Day(), t.Week(), t.Hour(), t.Minute())
}

// GetTimerID 获得标准化 ID
func (t *Timer) GetTimerID() uint32 {
	key := t.GetTimerInfo()
	m := md5.Sum(helper.StringToBytes(key))
	return binary.LittleEndian.Uint32(m[:4])
}

// GetFilledCronTimer 获得以cron填充好的ts
func GetFilledCronTimer(croncmd string, alert string, img string, botqq, gid int64) *Timer {
	var t Timer
	t.Alert = alert
	t.Cron = croncmd
	t.URL = img
	t.SelfID = botqq
	t.GrpID = gid
	return &t
}

// GetFilledTimer 获得填充好的ts
func GetFilledTimer(dateStrs []string, botqq, grp int64, matchDateOnly bool) *Timer {
	var t Timer
	var err error

	monthStr := []rune(dateStrs[1])
	dayWeekStr := []rune(dateStrs[2])
	hourStr := []rune(dateStrs[3])
	minuteStr := []rune(dateStrs[4])

	mon := time.Month(chineseNum2Int(monthStr))
	if err = validateMonth(mon); err != nil {
		t.Alert = err.Error()
		return &t
	}
	t.SetMonth(mon)

	if err = parseDayOrWeek(dayWeekStr, &t); err != nil {
		t.Alert = err.Error()
		return &t
	}

	if len(hourStr) == 3 {
		hourStr = []rune{hourStr[0], hourStr[2]}
	}
	h := chineseNum2Int(hourStr)
	if err = validateHour(h); err != nil {
		t.Alert = err.Error()
		return &t
	}
	t.SetHour(h)

	if len(minuteStr) == 3 {
		minuteStr = []rune{minuteStr[0], minuteStr[2]}
	}
	minute := chineseNum2Int(minuteStr)
	if err = validateMinute(minute); err != nil {
		t.Alert = err.Error()
		return &t
	}
	t.SetMinute(minute)

	if !matchDateOnly {
		if err = parseAdditionalFields(dateStrs, &t); err != nil {
			t.Alert = err.Error()
			return &t
		}
		t.SetEn(true)
	}

	t.SelfID = botqq
	t.GrpID = grp
	return &t
}

func parseDayOrWeek(dayWeekStr []rune, t *Timer) error {
	lenOfDW := len(dayWeekStr)

	switch {
	case lenOfDW == 4:
		dayWeekStr = []rune{dayWeekStr[0], dayWeekStr[2]}
		d := chineseNum2Int(dayWeekStr)
		if err := validateDay(d); err != nil {
			return err
		}
		t.SetDay(d)
	case dayWeekStr[lenOfDW-1] == rune('日'):
		dayWeekStr = dayWeekStr[:lenOfDW-1]
		d := chineseNum2Int(dayWeekStr)
		if err := validateDay(d); err != nil {
			return err
		}
		t.SetDay(d)
	case dayWeekStr[0] == rune('每'):
		t.SetWeek(-1)
	default:
		w := chineseNum2Int(dayWeekStr[1:])
		if w == 7 {
			w = 0
		}
		if err := validateWeek(time.Weekday(w)); err != nil {
			return err
		}
		t.SetWeek(time.Weekday(w))
	}
	return nil
}

func parseAdditionalFields(dateStrs []string, t *Timer) error {
	urlStr := dateStrs[5]
	if urlStr != "" {
		if len(urlStr) < 4 {
			return fmt.Errorf("url格式错误")
		}
		t.URL = urlStr[4:]
		logrus.Debugln("[群管]" + t.URL)
		if !validateURL(t.URL) {
			return fmt.Errorf("url非法")
		}
	}
	t.Alert = dateStrs[6]
	return nil
}

// chineseNum2Int 汉字数字转int，仅支持-10～99，最多两位数，其中"每"解释为-1，"每二"为-2，以此类推
func chineseNum2Int(rs []rune) int {
	r := -1
	l := len(rs)
	mai := rune('每')
	if unicode.IsDigit(rs[0]) { // 默认可能存在的第二位也为int
		r, _ = strconv.Atoi(string(rs))
	} else {
		switch {
		case rs[0] == mai:
			if l == 2 {
				r = -chineseChar2Int(rs[1])
			}
		case l == 1:
			r = chineseChar2Int(rs[0])
		default:
			ten := chineseChar2Int(rs[0])
			if ten != 10 {
				ten *= 10
			}
			ge := chineseChar2Int(rs[1])
			if ge == 10 {
				ge = 0
			}
			r = ten + ge
		}
	}
	return r
}

// chineseChar2Int 处理单个字符的映射0~10
func chineseChar2Int(c rune) int {
	if c == rune('日') || c == rune('天') { // 周日/周天
		return 7
	}
	match := []rune("零一二三四五六七八九十")
	for i, m := range match {
		if c == m {
			return i
		}
	}
	return 0
}
