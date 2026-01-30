package ports_scan

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	maxPortsPerScan = 1024
	defaultTimeout  = 3
	maxConcurrent   = 100
)

const (
	ProtocolTCP = "TCP"
	ProtocolUDP = "UDP"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "端口扫描",
		Help: "端口扫描\n" +
			"- /ports_scan <地址> - 扫描TCP常用端口\n" +
			"- /ports_scan <地址> -u - 扫描UDP常用端口\n" +
			"- /ports_scan <地址> -a - 扫描TCP和UDP常用端口\n" +
			"- /ports_scan <地址> -p <端口> - 扫描指定TCP端口\n" +
			"- /ports_scan <地址> -p <端口> -u - 扫描指定UDP端口\n" +
			"- /ports_scan <地址> -r <起始-结束> - 扫描TCP端口范围\n" +
			"- /ports_scan <地址> -r <起始-结束> -u - 扫描UDP端口范围\n" +
			"- /ports_scan <地址> -t <超时秒> - 设置超时时间\n" +
			"\nTCP常用端口: 21,22,23,25,80,110,143,443,465,587,993,995,3306,3389,5432,6379,8080,8443,27017\n" +
			"UDP常用端口: 53,67,68,69,123,161,162,514,520,4500,5000,5353,11211",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnRegex(`^/ports_scan\s+(.+?)(?:\s+-u)?(?:\s+-a)?(?:\s+-p\s+(\d+(?:,\d+)*))?(?:\s+-r\s+(\d+)-(\d+))?(?:\s+-t\s+(\d+))?(?:\s+-u)?(?:\s+-a)?$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			matches := ctx.State["regex_matched"].([]string)
			target := strings.TrimSpace(matches[1])

			if target == "" {
				ctx.SendChain(message.Text("错误: 请指定要扫描的地址\n使用 /ports_scan help 查看帮助"))
				return
			}

			input := matches[0]
			scanTCP := !strings.Contains(input, "-u") || strings.Contains(input, "-a")
			scanUDP := strings.Contains(input, "-u") || strings.Contains(input, "-a")

			timeout := defaultTimeout
			if matches[5] != "" {
				t, err := strconv.Atoi(matches[5])
				if err == nil && t > 0 && t <= 30 {
					timeout = t
				}
			}

			var ports []int

			if matches[2] != "" {
				portStrs := strings.Split(matches[2], ",")
				for _, ps := range portStrs {
					if p, err := strconv.Atoi(ps); err == nil && p > 0 && p <= 65535 {
						ports = append(ports, p)
					}
				}
			} else if matches[3] != "" && matches[4] != "" {
				start, err1 := strconv.Atoi(matches[3])
				end, err2 := strconv.Atoi(matches[4])
				if err1 == nil && err2 == nil && start > 0 && end <= 65535 && start <= end {
					if end-start+1 > maxPortsPerScan {
						ctx.SendChain(message.Text("错误: 端口范围过大，最多扫描 ", maxPortsPerScan, " 个端口"))
						return
					}
					for i := start; i <= end; i++ {
						ports = append(ports, i)
					}
				}
			} else {
				if scanTCP {
					ports = append(ports, getCommonTCPPorts()...)
				}
				if scanUDP {
					ports = append(ports, getCommonUDPPorts()...)
				}
			}

			if len(ports) == 0 {
				ctx.SendChain(message.Text("错误: 没有有效的端口"))
				return
			}

			if len(ports) > maxPortsPerScan {
				ctx.SendChain(message.Text("错误: 端口数量过多，最多扫描 ", maxPortsPerScan, " 个端口"))
				return
			}

			ctx.SendChain(message.Text("正在扫描 ", target, " 的 ", len(ports), " 个端口，超时 ", timeout, " 秒..."))

			results := scanPorts(target, ports, time.Duration(timeout)*time.Second, scanTCP, scanUDP)
			sendScanResults(ctx, target, results)
		})
}

func getCommonTCPPorts() []int {
	return []int{
		21,
		22,
		23,
		25,
		80,
		110,
		143,
		443,
		465,
		587,
		993,
		995,
		3306,
		3389,
		5432,
		6379,
		8080,
		8443,
		27017,
	}
}

func getCommonUDPPorts() []int {
	return []int{
		53,
		67,
		68,
		69,
		123,
		161,
		162,
		514,
		520,
		4500,
		5000,
		5353,
		11211,
	}
}

func scanPorts(target string, ports []int, timeout time.Duration, scanTCP, scanUDP bool) []PortResult {
	logrus.Infoln("[ports_scan] 扫描:", target, "端口数量:", len(ports), "TCP:", scanTCP, "UDP:", scanUDP)

	results := make([]PortResult, 0, len(ports)*2)
	resultsMutex := &sync.Mutex{}

	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, port := range ports {
		if scanTCP {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				open := isTCPPortOpen(target, p, timeout)
				resultsMutex.Lock()
				results = append(results, PortResult{
					Port:     p,
					Protocol: ProtocolTCP,
					Open:     open,
					Service:  getServiceName(p, ProtocolTCP),
				})
				resultsMutex.Unlock()
			}(port)
		}

		if scanUDP {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				open := isUDPPortOpen(target, p, timeout)
				resultsMutex.Lock()
				results = append(results, PortResult{
					Port:     p,
					Protocol: ProtocolUDP,
					Open:     open,
					Service:  getServiceName(p, ProtocolUDP),
				})
				resultsMutex.Unlock()
			}(port)
		}
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Port != results[j].Port {
			return results[i].Port < results[j].Port
		}
		return results[i].Protocol < results[j].Protocol
	})

	return results
}

func isTCPPortOpen(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func isUDPPortOpen(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)

	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return false
	}

	conn, err := net.DialTimeout("udp", udpAddr.String(), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	readTimeout := time.NewTimer(timeout)
	defer readTimeout.Stop()

	buffer := make([]byte, 1024)

	done := make(chan bool, 1)
	go func() {
		conn.SetReadDeadline(time.Now().Add(timeout))
		_, err := conn.Read(buffer)
		if err != nil {
			done <- false
		} else {
			done <- true
		}
	}()

	select {
	case result := <-done:
		return result
	case <-readTimeout.C:
		return true
	}
}

type PortResult struct {
	Port     int
	Protocol string
	Open     bool
	Service  string
}

func getServiceName(port int, protocol string) string {
	if protocol == ProtocolUDP {
		udpServices := map[int]string{
			53:    "DNS",
			67:    "DHCP-Server",
			68:    "DHCP-Client",
			69:    "TFTP",
			123:   "NTP",
			161:   "SNMP",
			162:   "SNMP-Trap",
			514:   "Syslog",
			520:   "RIP",
			4500:  "IPsec-NAT-T",
			5000:  "UPnP",
			5353:  "mDNS",
			11211: "Memcached",
		}

		if service, ok := udpServices[port]; ok {
			return service
		}
	} else {
		tcpServices := map[int]string{
			21:    "FTP",
			22:    "SSH",
			23:    "Telnet",
			25:    "SMTP",
			80:    "HTTP",
			110:   "POP3",
			143:   "IMAP",
			443:   "HTTPS",
			465:   "SMTPS",
			587:   "SMTP",
			993:   "IMAPS",
			995:   "POP3S",
			3306:  "MySQL",
			3389:  "RDP",
			5432:  "PostgreSQL",
			6379:  "Redis",
			8080:  "HTTP-Alt",
			8443:  "HTTPS-Alt",
			27017: "MongoDB",
		}

		if service, ok := tcpServices[port]; ok {
			return service
		}
	}
	return "Unknown"
}

func sendScanResults(ctx *zero.Ctx, target string, results []PortResult) {
	var openPorts []PortResult
	var closedPorts []PortResult

	for _, result := range results {
		if result.Open {
			openPorts = append(openPorts, result)
		} else {
			closedPorts = append(closedPorts, result)
		}
	}

	var sb strings.Builder
	sb.WriteString("=== 端口扫描结果: ")
	sb.WriteString(target)
	sb.WriteString(" ===\n\n")

	sb.WriteString(fmt.Sprintf("扫描端口: %d\n", len(results)))
	sb.WriteString(fmt.Sprintf("开放端口: %d\n", len(openPorts)))
	sb.WriteString(fmt.Sprintf("关闭端口: %d\n", len(closedPorts)))

	if len(openPorts) > 0 {
		sb.WriteString("\n开放端口:\n")
		for _, port := range openPorts {
			sb.WriteString(fmt.Sprintf("  %5d/%s  %s\n", port.Port, strings.ToLower(port.Protocol), port.Service))
		}
	} else {
		sb.WriteString("\n未发现开放端口\n")
	}

	if len(closedPorts) > 0 && len(closedPorts) <= 20 {
		sb.WriteString("\n关闭端口:\n")
		for _, port := range closedPorts {
			sb.WriteString(fmt.Sprintf("  %5d/%s  %s\n", port.Port, strings.ToLower(port.Protocol), port.Service))
		}
	} else if len(closedPorts) > 20 {
		sb.WriteString(fmt.Sprintf("\n其他 %d 个端口已关闭\n", len(closedPorts)))
	}

	sb.WriteString("注意: UDP端口扫描结果可能存在误报，因为某些服务可能不响应通用探测数据包\n")
	sb.WriteString("=====================")

	result := sb.String()
	if len(result) > 4000 {
		result = result[:4000] + "\n... (输出过长，已截断)"
	}

	ctx.SendChain(message.Text(result))
}
