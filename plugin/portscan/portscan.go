package portscan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type apiResponse struct {
	Code      int            `json:"code"`
	Msg       string         `json:"msg"`
	Data      map[string]bool `json:"data"`
	RequestID string         `json:"request_id"`
}

func init() {
	engine := control.Register("portscan", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "端口扫描工具",
		Help:             "- 端口扫描 [域名/IP]",
	})

	engine.OnCommand("端口扫描").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.Send("请输入要扫描的地址，例如：端口扫描 example.com")
			return
		}

		apiURL := fmt.Sprintf("https://v2.xxapi.cn/api/portscan?address=%s", args)

		resp, err := http.Get(apiURL)
		if err != nil {
			ctx.SendChain(message.Text("API请求失败：", err.Error()))
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ctx.SendChain(message.Text("读取响应失败：", err.Error()))
			return
		}

		var result apiResponse
		if err := json.Unmarshal(body, &result); err != nil {
			ctx.SendChain(message.Text("解析数据失败：", err.Error()))
			return
		}

		if result.Code != 200 {
			ctx.SendChain(message.Text("API错误：", result.Msg))
			return
		}

		// 构建合并转发消息（使用最新标准）
		msg := make(message.Message, 0, len(result.Data)+2)

		// 添加标题节点
		titleMsg := message.Message{
			message.Text(fmt.Sprintf("📡 扫描目标：%s", args)),
		}
		msg = append(msg, ctxext.FakeSenderForwardNode(ctx, titleMsg...))

		// 添加端口状态节点
		for port, status := range result.Data {
			state := "❌ 关闭"
			if status {
				state = "✅ 开放"
			}
			portMsg := message.Message{
				message.Text(fmt.Sprintf("端口 %-5s → %s", port, state)),
			}
			msg = append(msg, ctxext.FakeSenderForwardNode(ctx, portMsg...))
		}

		// 添加追踪信息节点
		traceMsg := message.Message{
			message.Text(fmt.Sprintf("🔖 请求ID：%s", result.RequestID)),
		}
		msg = append(msg, ctxext.FakeSenderForwardNode(ctx, traceMsg...))

		// 发送合并转发
		if id := ctx.Send(msg).ID(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 消息发送失败，可能被风控"))
		}
	})
}