// Package buckshotroulette 恶魔轮盘
package buckshot

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "恶魔轮盘",
		Help: "- 恶魔轮盘.创建游戏\n" +
			"- 恶魔轮盘.加入游戏\n" +
			"- 恶魔轮盘.开始游戏\n" +
			"- 恶魔轮盘.对战信息\n" +
			"- 恶魔轮盘.道具说明 [道具名]\n" +
			"- 恶魔轮盘.结束游戏\n" +
			"- 恶魔轮盘.设置禁言开启 [时长] (仅管理员)\n" +
			"- 恶魔轮盘.设置禁言关闭 (仅管理员)\n" +
			"- 恶魔轮盘.禁言设置 (仅管理员)\n" +
			"- 游戏中发送道具名使用道具\n" +
			"- 游戏中发送'自己'或'对方'开枪",
		PublicDataFolder: "BSR",
	})
)

func init() {
	itemList = getItemList()
	engine.OnPrefix("恶魔轮盘.创建游戏").SetBlock(true).Handle(handleCreateGame)
	engine.OnPrefix("恶魔轮盘.加入游戏").SetBlock(true).Handle(handleJoinGame)
	engine.OnPrefix("恶魔轮盘.开始游戏").SetBlock(true).Handle(handleStartGame)
	engine.OnPrefix("恶魔轮盘.对战信息").SetBlock(true).Handle(handleBattleInfo)
	engine.OnRegex(`^恶魔轮盘\.道具说明\s+(.+)$`).SetBlock(true).Handle(handleItemDescription)
	engine.OnPrefix("恶魔轮盘.结束游戏").SetBlock(true).Handle(handleEndGame)
	engine.OnRegex(`^(自己|对方)$`).SetBlock(true).Handle(handleShoot)
	engine.OnMessage().SetBlock(false).Handle(handleUseItem)
	engine.OnPrefix("恶魔轮盘.设置禁言").SetBlock(true).Handle(handleSetMuteConfig)
	engine.OnPrefix("恶魔轮盘.禁言设置").SetBlock(true).Handle(handleSetMuteConfig)
}

func isAdmin(ctx *zero.Ctx) bool {
	return zero.AdminPermission(ctx)
}
