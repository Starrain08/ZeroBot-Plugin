package timer

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	maxMonth   = 12
	maxDay     = 31
	maxHour    = 23
	maxMinute  = 59
	maxWeekDay = 6
)

func validateMonth(month time.Month) error {
	if month < -1 || month > maxMonth {
		return fmt.Errorf("月份非法！")
	}
	return nil
}

func validateDay(day int) error {
	if day < -1 || day > maxDay {
		return fmt.Errorf("日期非法！")
	}
	return nil
}

func validateWeek(week time.Weekday) error {
	if week < -1 || week > maxWeekDay {
		return fmt.Errorf("星期非法！")
	}
	return nil
}

func validateHour(hour int) error {
	if hour < -1 || hour > maxHour {
		return fmt.Errorf("小时非法！")
	}
	return nil
}

func validateMinute(minute int) error {
	if minute < -1 || minute > maxMinute {
		return fmt.Errorf("分钟非法！")
	}
	return nil
}

func validateURL(url string) bool {
	if url == "" {
		return true
	}
	valid := len(url) >= 4 && (url[:4] == "http")
	if !valid {
		logrus.Debugln("[群管]url非法！")
	}
	return valid
}
