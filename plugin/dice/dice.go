// Package dice dice！golang部分移植版
package dice

import (
	"time"

	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	db     sql.Sqlite
	engine = control.Register("dice", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "骰子?",
		Help:             "试图移植的dice\n-.jrrp\n-.ra\n-.rd",
		PublicDataFolder: "Dice",
	})
)

func init() {
	go func() {
		db = sql.New(engine.DataFolder() + "dice.db")
		if err := db.Open(time.Hour * 24); err != nil {
			panic(err)
		}
		tables := []struct {
			name  string
			model interface{}
		}{
			{"strjrrp", &strjrrp{}},
			{"rsl", &rsl{}},
			{"set", &set{}},
		}
		for _, t := range tables {
			if err := db.Create(t.name, t.model); err != nil {
				panic(err)
			}
		}
	}()
}
