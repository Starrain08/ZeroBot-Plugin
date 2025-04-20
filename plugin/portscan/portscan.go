package portscan

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "net/http"

    "github.com/FloatTech/zbpctrl"
    "github.com/FloatTech/zbputils/control"
    "github.com/wdvxdr1123/ZeroBot/message"
    zero "github.com/wdvxdr1123/ZeroBot"
)

// API地址
const portscanAPI = "https://v2.xxapi.cn/api/portscan"

func init() {
    engine := control.Register("portscan", &ctrl.Options[*zero.Ctx]{
        DisableOnDefault: false,
        Brief:            "端口扫描插件",
        Help:             "- 端口扫描 [IP地址]（仅管理员可用）检查指定IP的常用端口开放情况",
        PublicDataFolder: "PortScan",
    })

    // 创建命令触发器，仅管理员可触发
    engine.OnCommand("端口扫描", zero.AdminPermission).SetBlock(true).
        Handle(func(ctx *zero.Ctx) {
            // 权限二次验证（保险）
            if !zero.CheckAdmin(ctx) {
                ctx.SendChain(message.Text("权限不足，仅管理员可操作"))
                return
            }

            // 解析参数
            args := ctx.NormArgs()
            if len(args) < 1 {
                ctx.SendChain(message.Text("用法：端口扫描 [IP地址]"))
                return
            }
            ip := args[0]

            // 调用API
            resp, err := http.Get(portscanAPI + "?address=" + ip)
            if err != nil {
                ctx.SendChain(message.Text("请求失败: " + err.Error()))
                return
            }
            defer resp.Body.Close()

            // 解析响应
            body, _ := ioutil.ReadAll(resp.Body)
            var result struct {
                Code int            `json:"code"`
                Data map[string]bool `json:"data"`
                Msg  string         `json:"msg"`
            }
            json.Unmarshal(body, &result)

            // 处理API返回状态
            if result.Code != 200 {
                ctx.SendChain(message.Text("API错误: " + result.Msg))
                return
            }

            // 提取开放端口
            var openPorts []string
            for port, isOpen := range result.Data {
                if isOpen {
                    openPorts = append(openPorts, port)
                }
            }

            // 构建合并转发消息
            nodes := []message.Node{
                {
                    User: "端口扫描助手",
                    Content: message.Text(
                        "IP地址:", ip,
                        "\n\n扫描结果：",
                    ),
                },
            }

            if len(openPorts) == 0 {
                nodes[0].Content = message.Text("未发现开放的常用端口")
            } else {
                nodes = append(nodes, message.Node{
                    User: "端口扫描助手",
                    Content: message.Text(
                        "开放端口列表：",
                        "\n" + bytes.TrimSuffix(bytes.Join(
                            bytes.Split([]byte(strings.Join(openPorts, "\n")), '\n'),
                            "\n"), nil),
                    ),
                })
            }

            // 发送合并消息
            ctx.Send(message.Nodes(nodes...))
        })
}