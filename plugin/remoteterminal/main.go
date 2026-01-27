package remoteterminal

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/sirupsen/logrus"
)

const (
	maxOutputLines = 30
	maxLineLength = 200
	defaultTimeout = 30
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "远程终端管理",
		Help:             "远程终端管理\n" +
			"- /terminal exec <命令> - 执行命令\n" +
			"- /terminal cd <路径> - 切换工作目录\n" +
			"- /terminal pwd - 显示当前目录\n" +
			"- /terminal ls - 列出当前目录文件\n" +
			"- /terminal timeout <秒> - 设置命令超时时间（默认30秒）\n" +
			"- /terminal help - 显示帮助",
	}).ApplySingle(ctxext.DefaultSingle)

	engine, _ := control.Lookup("remoteterminal")

	var currentDir string
	cmdTimeout := defaultTimeout

	engine.OnPrefix("/terminal", zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			parts := strings.Fields(args)
			
			if len(parts) == 0 {
				ctx.SendChain(message.Text(getHelp()))
				return
			}

			cmd := parts[0]

			switch cmd {
			case "exec":
				if len(parts) < 2 {
					ctx.SendChain(message.Text("错误: 请指定要执行的命令\n使用 /terminal help 查看帮助"))
					return
				}
				command := strings.Join(parts[1:], " ")
				output, err := executeCommand(command, currentDir, time.Duration(cmdTimeout)*time.Second)
				if err != nil {
					ctx.SendChain(message.Text("执行失败: ", err.Error()))
				} else {
					sendOutput(ctx, output)
				}
			
			case "cd":
				if len(parts) < 2 {
					ctx.SendChain(message.Text("错误: 请指定目标目录"))
					return
				}
				path := strings.Join(parts[1:], " ")
				newDir, err := changeDirectory(path, currentDir)
				if err != nil {
					ctx.SendChain(message.Text("切换目录失败: ", err.Error()))
				} else {
					currentDir = newDir
					ctx.SendChain(message.Text("当前目录: ", currentDir))
				}
			
			case "pwd":
				output, err := executeCommand("pwd", currentDir, 5*time.Second)
				if err != nil {
					ctx.SendChain(message.Text("获取当前目录失败: ", err.Error()))
				} else {
					ctx.SendChain(message.Text(strings.TrimSpace(output)))
				}
			
			case "ls":
				output, err := executeCommand("ls -la", currentDir, 10*time.Second)
				if err != nil {
					ctx.SendChain(message.Text("列出文件失败: ", err.Error()))
				} else {
					sendOutput(ctx, output)
				}
			
			case "timeout":
				if len(parts) < 2 {
					ctx.SendChain(message.Text("当前超时设置: ", cmdTimeout, " 秒\n使用 /terminal timeout <秒> 设置超时时间"))
					return
				}
				var timeout int
				_, err := fmt.Sscanf(parts[1], "%d", &timeout)
				if err != nil || timeout <= 0 || timeout > 300 {
					ctx.SendChain(message.Text("错误: 超时时间必须是1-300之间的整数"))
					return
				}
				cmdTimeout = timeout
				ctx.SendChain(message.Text("命令超时已设置为 ", cmdTimeout, " 秒"))
			
			case "help":
				ctx.SendChain(message.Text(getHelp()))
			
			default:
				ctx.SendChain(message.Text("未知命令: ", cmd, "\n使用 /terminal help 查看帮助"))
			}
		})
}

func executeCommand(command, dir string, timeout time.Duration) (string, error) {
	logrus.Infoln("[remoteterminal] 执行命令:", command, "在目录:", dir)
	
	cmd := exec.Command("sh", "-c", command)
	if dir != "" {
		cmd.Dir = dir
	}
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Run()
	}()
	
	select {
	case err := <-errChan:
		if err != nil {
			if stderr.Len() > 0 {
				return stderr.String(), err
			}
			return "", err
		}
		if stderr.Len() > 0 {
			return stdout.String() + "\n[stderr]\n" + stderr.String(), nil
		}
		return stdout.String(), nil
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "", fmt.Errorf("命令执行超时（超过 %v 秒）", timeout.Seconds())
	}
}

func changeDirectory(path, currentDir string) (string, error) {
	var newDir string
	
	if strings.HasPrefix(path, "/") {
		newDir = path
	} else if currentDir == "" {
		newDir = path
	} else {
		newDir = currentDir + "/" + path
	}
	
	cmd := exec.Command("sh", "-c", "cd "+newDir+" && pwd")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	
	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", fmt.Errorf("无法访问目录: %s", newDir)
	}
	
	return result, nil
}

func sendOutput(ctx *zero.Ctx, output string) {
	lines := strings.Split(output, "\n")
	
	if len(lines) <= maxOutputLines {
		result := strings.Join(lines, "\n")
		if len(result) > 4000 {
			result = result[:4000] + "\n... (输出过长，已截断)"
		}
		ctx.SendChain(message.Text(result))
		return
	}
	
	truncated := make([]string, maxOutputLines+1)
	truncated[0] = fmt.Sprintf("输出共 %d 行，显示前 %d 行:", len(lines), maxOutputLines)
	copy(truncated[1:], lines[:maxOutputLines])
	
	result := strings.Join(truncated, "\n")
	if len(result) > 4000 {
		result = result[:4000] + "\n... (输出过长，已截断)"
	}
	
	ctx.SendChain(message.Text(result))
}

func getHelp() string {
	return `=== 远程终端管理帮助 ===

可用命令:
/terminal exec <命令>     - 执行 shell 命令
/terminal cd <路径>       - 切换工作目录
/terminal pwd             - 显示当前工作目录
/terminal ls              - 列出当前目录文件
/terminal timeout <秒>    - 设置命令超时时间 (1-300秒，默认30秒)
/terminal help            - 显示此帮助信息

注意:
- 仅超级用户可使用
- 命令执行有超时限制
- 输出过长会被截断
- 支持基本的 shell 命令

示例:
/terminal exec ls -la
/terminal exec docker ps
/terminal exec python --version
/terminal timeout 60
`
}
