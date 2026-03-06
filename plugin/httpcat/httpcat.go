package httpcat

import (
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "HTTP状态码猫图",
		Help:             "- http <状态码>\n例: http 404",
	})
	engine.OnPrefix("http").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		statusCodeStr := strings.TrimSpace(ctx.State["args"].(string))
		statusCode, err := strconv.Atoi(statusCodeStr)
		if err != nil {
			ctx.SendChain(message.Text("请输入正确的HTTP状态码,例如: http 404"))
			return
		}
		url := "https://http.cat/" + statusCodeStr
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			ctx.SendChain(message.Text("获取图片失败: ", err))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				ctx.SendChain(message.Text("读取图片失败: ", err))
				return
			}
			base64Img := base64.StdEncoding.EncodeToString(body)
			ctx.SendChain(message.Image("base64://" + base64Img))
		} else {
			ctx.SendChain(message.Text("你家 http 协议会返回 ", statusCode, "？"))
		}
	})
}
