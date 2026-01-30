package remoteterminal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
	maxOutputLines = 30
	maxLineLength  = 200
	defaultTimeout = 30
	blacklistFile  = "blacklist.txt"
)

type dangerousCommand struct {
	pattern string
	reason  string
}

var (
	dangerousCommands []dangerousCommand
	isWindows         bool
	dataFolder        string
)

func getDefaultLinuxBlacklist() []dangerousCommand {
	return []dangerousCommand{
		{`^(rm\s+.*-rf\s+/)|^(rm\s+-rf\s+.*)`, "删除系统文件"},
		{`^dd\s+`, "磁盘写入操作"},
		{`^mkfs\.?\w*\s+`, "格式化文件系统"},
		{`^(shutdown|reboot|poweroff|halt)`, "系统关机/重启"},
		{`^(killall\s+-9)|(kill\s+-9)`, "强制终止进程"},
		{`^(chmod|chown)\s+-R\s+/`, "递归修改系统权限"},
		{`^(useradd|userdel|usermod)\s+`, "用户管理"},
		{`^(groupadd|groupdel)\s+`, "组管理"},
		{`^(fdisk|parted)\s+`, "磁盘分区操作"},
		{`^(iptables|ufw|firewall-cmd)\s+`, "防火墙规则修改"},
		{`crontab\s+.*-(e|r)`, "编辑或删除定时任务"},
		{`(systemctl|service)\s+.*(stop|restart|disable)`, "停止/重启系统服务"},
		{`^init\s+\d+`, "切换系统运行级别"},
		{`^(su|sudo)\s+`, "权限提升"},
		{`^passwd\s+`, "修改密码"},
		{`^:\s+.*>\s*/`, "重定向写入系统文件"},
		{`^echo\s+.*>>\s+/`, "追加写入系统文件"},
		{`^mv\s+.*\s+/`, "移动文件到系统目录"},
		{`^cp\s+.*\s+/`, "复制文件到系统目录"},
	}
}

func getDefaultWindowsBlacklist() []dangerousCommand {
	return []dangerousCommand{
		{`(?i)^(del\s+.*[\\/](windows|system|program files))`, "删除系统文件"},
		{`(?i)^(rd\s+.*[\\/](windows|system|program files))`, "删除系统目录"},
		{`(?i)^(rmdir\s+.*[\\/](windows|system|program files))`, "删除系统目录"},
		{`(?i)^(erase\s+.*[\\/](windows|system|program files))`, "删除系统文件"},
		{`(?i)^format\s+[a-z]:`, "格式化磁盘"},
		{`(?i)^(shutdown|restart|logoff)`, "系统关机/重启/注销"},
		{`(?i)^taskkill\s+.*/f`, "强制终止进程"},
		{`(?i)^net\s+(user|group)\s+.*/delete`, "删除用户/组"},
		{`(?i)^net\s+share\s+.*/delete`, "删除共享"},
		{`(?i)^diskpart`, "磁盘分区操作"},
		{`(?i)^bcdedit`, "启动配置编辑"},
		{`(?i)^reg\s+delete\s+.*HKEY_LOCAL_MACHINE`, "删除注册表项"},
		{`(?i)^sc\s+(stop|delete|config)`, "停止/删除系统服务"},
		{`(?i)^icacls\s+.*[\\/](windows|system)`, "修改系统权限"},
		{`(?i)^(move|xcopy)\s+.*[\\/](windows|system|program files)`, "移动/复制系统文件"},
		{`(?i)^echo\s+.*>\s*[A-Z]:[\\/]windows`, "写入系统文件"},
		{`(?i)^cmd\s+/c\s+.*(del|format|rd|rmdir)`, "执行危险命令"},
		{`(?i)^powershell\s+.*(Remove-Item|Stop-Computer|Restart-Computer)`, "执行危险PowerShell命令"},
	}
}

func loadCustomBlacklist(folder string) error {
	if err := os.MkdirAll(folder, 0755); err != nil {
		logrus.Errorf("[remoteterminal] 创建数据文件夹失败: %v", err)
		return err
	}

	filePath := filepath.Join(folder, blacklistFile)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Infoln("[remoteterminal] 未找到自定义黑名单文件，使用默认黑名单")
			if err := createBlacklistTemplate(filePath); err != nil {
				logrus.Warnf("[remoteterminal] 创建黑名单模板文件失败: %v", err)
			}
			return nil
		}
		return err
	}
	defer file.Close()

	logrus.Infoln("[remoteterminal] 加载自定义黑名单:", filePath)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			logrus.Warnf("[remoteterminal] 黑名单第 %d 行格式错误，跳过: %s", lineNum, line)
			continue
		}

		pattern := strings.TrimSpace(parts[0])
		reason := strings.TrimSpace(parts[1])

		if pattern == "" || reason == "" {
			logrus.Warnf("[remoteterminal] 黑名单第 %d 行格式错误，跳过: %s", lineNum, line)
			continue
		}

		dangerousCommands = append(dangerousCommands, dangerousCommand{
			pattern: pattern,
			reason:  reason,
		})

		logrus.Debugf("[remoteterminal] 添加自定义黑名单: %s -> %s", pattern, reason)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	logrus.Infof("[remoteterminal] 已加载 %d 条自定义黑名单规则", len(dangerousCommands)-len(getDefaultBlacklist()))
	return nil
}

func getDefaultBlacklist() []dangerousCommand {
	if isWindows {
		return getDefaultWindowsBlacklist()
	}
	return getDefaultLinuxBlacklist()
}

func createBlacklistTemplate(filePath string) error {
	content := `# 远程终端危险命令黑名单配置文件
# 格式: 正则表达式|原因说明
# 每行一条规则，以 # 开头的行为注释
#
# 示例:
# ^rm\s+-rf|删除文件命令
# ^shutdown|关机命令
# ^format\s+[a-z]:|格式化磁盘
`
	return os.WriteFile(filePath, []byte(content), 0644)
}

func init() {
	isWindows = runtime.GOOS == "windows"

	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "远程终端管理",
		Help: "远程终端管理\n" +
			"- /terminal exec <命令> - 执行命令\n" +
			"- /terminal cd <路径> - 切换工作目录\n" +
			"- /terminal pwd - 显示当前目录\n" +
			"- /terminal ls - 列出当前目录文件\n" +
			"- /terminal timeout <秒> - 设置命令超时时间（默认30秒）\n" +
			"- /terminal reload - 重新加载黑名单配置\n" +
			"- /terminal list_blacklist - 列出当前黑名单\n" +
			"- /terminal help - 显示帮助",
	}).ApplySingle(ctxext.DefaultSingle)

	dataFolder = engine.DataFolder()

	dangerousCommands = getDefaultBlacklist()

	go func() {
		if err := loadCustomBlacklist(dataFolder); err != nil {
			logrus.Errorf("[remoteterminal] 加载自定义黑名单失败: %v", err)
		}
	}()

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

			case "reload":
				dangerousCommands = getDefaultBlacklist()
				if err := loadCustomBlacklist(dataFolder); err != nil {
					ctx.SendChain(message.Text("重新加载黑名单失败: ", err.Error()))
				} else {
					ctx.SendChain(message.Text("黑名单已重新加载，当前共 ", len(dangerousCommands), " 条规则"))
				}

			case "list_blacklist":
				listBlacklist(ctx)

			case "help":
				ctx.SendChain(message.Text(getHelp()))

			default:
				ctx.SendChain(message.Text("未知命令: ", cmd, "\n使用 /terminal help 查看帮助"))
			}
		})
}

func listBlacklist(ctx *zero.Ctx) {
	var sb strings.Builder

	if isWindows {
		sb.WriteString("=== Windows 危险命令黑名单 ===\n\n")
	} else {
		sb.WriteString("=== Linux 危险命令黑名单 ===\n\n")
	}

	sb.WriteString(fmt.Sprintf("当前操作系统: %s\n", runtime.GOOS))
	sb.WriteString(fmt.Sprintf("黑名单规则总数: %d\n\n", len(dangerousCommands)))

	defaultCount := len(getDefaultBlacklist())
	customCount := len(dangerousCommands) - defaultCount

	sb.WriteString(fmt.Sprintf("默认规则: %d 条\n", defaultCount))
	sb.WriteString(fmt.Sprintf("自定义规则: %d 条\n\n", customCount))

	sb.WriteString("规则列表:\n")
	for i, dc := range dangerousCommands {
		if i >= 50 {
			sb.WriteString(fmt.Sprintf("\n... (还有 %d 条规则未显示)", len(dangerousCommands)-50))
			break
		}
		prefix := "[默认] "
		if i >= defaultCount && customCount > 0 {
			prefix = "[自定义] "
		}
		sb.WriteString(fmt.Sprintf("%d. %s%s -> %s\n", i+1, prefix, dc.pattern, dc.reason))
	}

	sb.WriteString("\n=====================\n")
	sb.WriteString("自定义黑名单配置文件:\n")
	sb.WriteString(fmt.Sprintf("位置: %s\n", filepath.Join(dataFolder, blacklistFile)))
	sb.WriteString("格式: 正则表达式|原因说明\n")
	sb.WriteString("示例:\n")
	sb.WriteString("^rm\\s+-rf|删除文件\n")
	sb.WriteString("^shutdown|关机命令\n")

	result := sb.String()
	if len(result) > 4000 {
		result = result[:4000] + "\n... (内容过长，已截断)"
	}

	ctx.SendChain(message.Text(result))
}

func isDangerousCommand(command string) (bool, string) {
	trimmedCmd := strings.TrimSpace(command)

	for _, dc := range dangerousCommands {
		matched, _ := regexp.MatchString(dc.pattern, trimmedCmd)
		if matched {
			return true, dc.reason
		}
	}

	return false, ""
}

func executeCommand(command, dir string, timeout time.Duration) (string, error) {
	if dangerous, reason := isDangerousCommand(command); dangerous {
		return "", fmt.Errorf("危险命令被拦截: %s (原因: %s)", command, reason)
	}

	logrus.Infoln("[remoteterminal] 执行命令:", command, "在目录:", dir)

	var cmd *exec.Cmd
	if isWindows {
		cmd = exec.Command("cmd", "/c", "chcp 65001 >nul 2>&1 && "+command)
		cmd.Env = append(os.Environ(), "LANG=zh_CN.UTF-8", "LC_ALL=zh_CN.UTF-8")
	} else {
		cmd = exec.Command("sh", "-c", command)
		cmd.Env = append(os.Environ(), "LANG=zh_CN.UTF-8", "LC_ALL=zh_CN.UTF-8")
	}

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

	if isWindows {
		if len(path) >= 2 && strings.ToUpper(path[0:2]) == ":" {
			newDir = path
		} else if len(path) >= 1 && path[0] == '/' {
			if len(currentDir) >= 2 && currentDir[1] == ':' {
				newDir = string(currentDir[0]) + ":" + path
			} else {
				newDir = path
			}
		} else if currentDir == "" {
			newDir = path
		} else {
			if !strings.HasSuffix(currentDir, "\\") && !strings.HasSuffix(currentDir, "/") {
				newDir = currentDir + "\\" + path
			} else {
				newDir = currentDir + path
			}

			cmd := exec.Command("cmd", "/c", "chcp 65001 >nul 2>&1 && cd /d "+newDir+" && cd")
			cmd.Env = append(os.Environ(), "LANG=zh_CN.UTF-8", "LC_ALL=zh_CN.UTF-8")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			err := cmd.Run()
			if err != nil {
				return "", err
			}
		}

		cmd := exec.Command("cmd", "/c", "cd /d "+newDir+" && cd")
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
	} else {
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
	if isWindows {
		return `=== 远程终端管理帮助 (Windows) ===

可用命令:
/terminal exec <命令>         - 执行 Windows 命令
/terminal cd <路径>           - 切换工作目录
/terminal pwd                 - 显示当前工作目录
/terminal dir                 - 列出当前目录文件
/terminal timeout <秒>        - 设置命令超时时间 (1-300秒，默认30秒)
/terminal reload              - 重新加载黑名单配置
/terminal list_blacklist      - 列出当前危险命令黑名单
/terminal help                - 显示此帮助信息

安全限制:
- 仅超级用户可使用
- 命令执行有超时限制
- 输出过长会被截断
- 禁止执行危险命令，包括:
  * del/erase (删除系统文件)
  * rd/rmdir (删除系统目录)
  * format (格式化磁盘)
  * shutdown/restart (系统关机/重启)
  * taskkill /f (强制终止进程)
  * net user/group delete (用户/组管理)
  * diskpart (磁盘分区操作)
  * bcdedit (启动配置编辑)
  * reg delete (删除注册表项)
  * sc stop/delete (停止/删除系统服务)
  * 其他系统级危险操作

自定义黑名单:
- 在插件数据目录创建 blacklist.txt 文件
- 文件格式: 正则表达式|原因说明
- 每行一条规则，支持 # 开头的注释
- 使用 /terminal reload 重新加载配置

示例:
/terminal exec dir
/terminal exec ipconfig
/terminal exec tasklist
/terminal exec netstat -an
/terminal timeout 60
`
	} else {
		return `=== 远程终端管理帮助 (Linux) ===

可用命令:
/terminal exec <命令>         - 执行 shell 命令
/terminal cd <路径>           - 切换工作目录
/terminal pwd                 - 显示当前工作目录
/terminal ls                  - 列出当前目录文件
/terminal timeout <秒>        - 设置命令超时时间 (1-300秒，默认30秒)
/terminal reload              - 重新加载黑名单配置
/terminal list_blacklist      - 列出当前危险命令黑名单
/terminal help                - 显示此帮助信息

安全限制:
- 仅超级用户可使用
- 命令执行有超时限制
- 输出过长会被截断
- 禁止执行危险命令，包括:
  * rm -rf (强制删除系统文件)
  * dd (磁盘写入操作)
  * mkfs (格式化文件系统)
  * shutdown/reboot/poweroff (系统关机/重启)
  * kill -9 / killall -9 (强制终止进程)
  * chmod/chown -R / (递归修改系统权限)
  * useradd/userdel (用户管理)
  * fdisk/parted (磁盘分区)
  * iptables/ufw (防火墙规则修改)
  * systemctl service stop/disable (停止系统服务)
  * su/sudo (权限提升)
  * 其他系统级危险操作

自定义黑名单:
- 在插件数据目录创建 blacklist.txt 文件
- 文件格式: 正则表达式|原因说明
- 每行一条规则，支持 # 开头的注释
- 使用 /terminal reload 重新加载配置

示例:
/terminal exec ls -la
/terminal exec docker ps
/terminal exec python --version
/terminal exec tail -f /var/log/syslog
/terminal timeout 60
`
	}
}
