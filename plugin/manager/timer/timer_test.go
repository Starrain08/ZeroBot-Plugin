package timer

import (
	"testing"
	"time"

	sql "github.com/FloatTech/sqlite"
	"github.com/sirupsen/logrus"
)

func TestNextWakeTime(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ts := &Timer{}
	ts.SetMonth(-1)
	ts.SetWeek(6)
	ts.SetHour(16)
	ts.SetMinute(30)

	nextTime := ts.nextWakeTime()
	duration := time.Until(nextTime)

	if duration <= 0 {
		t.Fatalf("预期唤醒时间在未来，但得到: %v", nextTime)
	}

	t.Logf("下次唤醒时间: %v, 距离现在: %v", nextTime, duration)
}

func TestTimerOperations(t *testing.T) {
	db := sql.New(":memory:")
	defer db.Close()

	c := NewClock(&db)
	defer c.Shutdown()

	testTimer := GetFilledTimer([]string{"", "12", "-1", "12", "0", "", "test alarm"}, 0, 12345, false)
	testTimer.SetEn(false)

	err := c.AddTimerIntoDB(testTimer)
	if err != nil {
		t.Fatalf("添加定时器失败: %v", err)
	}

	list := c.ListTimers(12345)
	if len(list) == 0 {
		t.Error("未找到测试定时器")
	}
	t.Logf("定时器列表: %v", list)

	key := testTimer.GetTimerID()
	ok := c.CancelTimer(key)
	if !ok {
		t.Error("取消定时器失败")
	}
}

func TestChineseNum2Int(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"一", 1},
		{"十", 10},
		{"十二", 12},
		{"二十", 20},
		{"九十九", 99},
		{"零", 0},
	}

	for _, tt := range tests {
		result := chineseNum2Int([]rune(tt.input))
		if result != tt.expected {
			t.Errorf("chineseNum2Int(%q) = %d, 期望 %d", tt.input, result, tt.expected)
		}
	}
}
