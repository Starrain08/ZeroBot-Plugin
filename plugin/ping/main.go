package ping

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	defaultCount   = 4
	defaultTimeout = 5
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "Ping 检测",
		Help: "Ping 检测\n" +
			"- /ping <地址> - Ping 指定地址（默认4次）\n" +
			"- /ping <地址> -c <次数> - Ping 指定次数\n" +
			"- /ping <地址> -t <超时秒数> - 设置超时时间\n" +
			"- /ping <地址> -c <次数> -t <超时秒数> - 指定次数和超时\n",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnRegex(`^/ping\s+(.+?)(?:\s+-c\s+(\d+))?(?:\s+-t\s+(\d+))?$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			matches := ctx.State["regex_matched"].([]string)
			target := strings.TrimSpace(matches[1])

			count := defaultCount
			if matches[2] != "" {
				c, err := strconv.Atoi(matches[2])
				if err == nil && c > 0 && c <= 100 {
					count = c
				}
			}

			timeout := defaultTimeout
			if matches[3] != "" {
				t, err := strconv.Atoi(matches[3])
				if err == nil && t > 0 && t <= 300 {
					timeout = t
				}
			}

			ctx.SendChain(message.Text("正在 Ping ", target, " (", count, " 次，超时 ", timeout, " 秒)..."))

			output, err := doPing(target, count, timeout)
			if err != nil {
				ctx.SendChain(message.Text("Ping 失败: ", err.Error()))
				return
			}

			sendPingResult(ctx, target, output)
		})
}

func doPing(target string, count int, timeout int) (string, error) {
	logrus.Infoln("[ping] 开始 ping:", target, "次数:", count, "超时:", timeout)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), "-w", strconv.Itoa(timeout*1000), target)
	case "darwin", "freebsd", "netbsd":
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-W", strconv.Itoa(timeout), target)
	default:
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-w", strconv.Itoa(timeout), target)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			if stderr.Len() > 0 {
				return stderr.String(), err
			}
			return stdout.String(), err
		}
		return stdout.String(), nil
	case <-time.After(time.Duration(timeout)*time.Second + 10*time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "", fmt.Errorf("ping 命令执行超时")
	}
}

func sendPingResult(ctx *zero.Ctx, target string, output string) {
	lines := strings.Split(output, "\n")

	var result []string
	result = append(result, "=== Ping 结果: "+target+" ===")

	totalSent := 0
	totalReceived := 0
	totalLoss := 0
	minTime := float64(-1)
	maxTime := float64(-1)
	sumTime := 0.0
	timeCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result = append(result, line)

		if runtime.GOOS == "windows" {
			if strings.Contains(line, "已发送") || strings.Contains(line, "Sent") {
				parts := strings.Fields(line)
				for i, p := range parts {
					if strings.Contains(p, "=") && i < len(parts)-1 {
						val, err := strconv.Atoi(strings.TrimSuffix(parts[i+1], ","))
						if err == nil {
							if strings.Contains(p, "发送") || strings.Contains(p, "Sent") {
								totalSent = val
							}
						}
					}
				}
			}
			if strings.Contains(line, "接收") || strings.Contains(line, "Received") {
				parts := strings.Fields(line)
				for i, p := range parts {
					if strings.Contains(p, "=") && i < len(parts)-1 {
						val, err := strconv.Atoi(strings.TrimSuffix(parts[i+1], ","))
						if err == nil {
							if strings.Contains(p, "接收") || strings.Contains(p, "Received") {
								totalReceived = val
							}
						}
					}
				}
			}

			if strings.Contains(line, "ms") && strings.Contains(line, "<") && strings.Contains(line, "time") {
				timeStr := line
				start := strings.LastIndex(timeStr, "=") + 1
				end := strings.Index(timeStr, "ms")
				if start > 0 && end > start {
					timeValStr := timeStr[start:end]
					timeVal, err := strconv.ParseFloat(strings.TrimSpace(timeValStr), 64)
					if err == nil {
						if minTime < 0 || timeVal < minTime {
							minTime = timeVal
						}
						if maxTime < 0 || timeVal > maxTime {
							maxTime = timeVal
						}
						sumTime += timeVal
						timeCount++
					}
				}
			}
		} else {
			if strings.Contains(line, "packets transmitted") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					if s, err := strconv.Atoi(parts[0]); err == nil {
						totalSent = s
					}
					if s, err := strconv.Atoi(parts[3]); err == nil {
						totalReceived = s
					}
					if strings.Contains(line, "received") && len(parts) >= 6 {
						if s, err := strconv.Atoi(strings.TrimSuffix(parts[5], "%")); err == nil {
							totalLoss = s
						}
					}
				}
			}

			if strings.Contains(line, "min/avg/max") || strings.Contains(line, "rtt min/avg/max") {
				parts := strings.Fields(line)
				for _, p := range parts {
					if strings.Contains(p, "/") {
						times := strings.Split(p, "/")
						if len(times) >= 4 {
							if t, err := strconv.ParseFloat(times[0], 64); err == nil {
								minTime = t
							}
							if t, err := strconv.ParseFloat(times[1], 64); err == nil {
								sumTime = t * float64(timeCount)
							}
							if t, err := strconv.ParseFloat(times[2], 64); err == nil {
								maxTime = t
							}
						}
					}
				}
			}

			if strings.Contains(line, "time=") && strings.Contains(line, "ms") {
				start := strings.Index(line, "time=") + 5
				end := strings.Index(line[start:], " ")
				if end > 0 {
					timeValStr := line[start : start+end]
					timeVal, err := strconv.ParseFloat(timeValStr, 64)
					if err == nil {
						if minTime < 0 || timeVal < minTime {
							minTime = timeVal
						}
						if maxTime < 0 || timeVal > maxTime {
							maxTime = timeVal
						}
						sumTime += timeVal
						timeCount++
					}
				}
			}
		}
	}

	if totalSent > 0 {
		if totalLoss == 0 {
			totalLoss = totalSent - totalReceived
		}
		lossRate := float64(totalLoss) / float64(totalSent) * 100

		result = append(result, "")
		result = append(result, fmt.Sprintf("发送: %d  接收: %d  丢失: %d  丢失率: %.1f%%",
			totalSent, totalReceived, totalLoss, lossRate))

		if timeCount > 0 && minTime >= 0 && maxTime >= 0 {
			avgTime := sumTime / float64(timeCount)
			result = append(result, fmt.Sprintf("延迟: 最小 %.1fms  平均 %.1fms  最大 %.1fms",
				minTime, avgTime, maxTime))
		}
	}

	result = append(result, "=====================")

	finalOutput := strings.Join(result, "\n")
	if len(finalOutput) > 4000 {
		finalOutput = finalOutput[:4000] + "\n... (输出过长，已截断)"
	}

	ctx.SendChain(message.Text(finalOutput))
}
