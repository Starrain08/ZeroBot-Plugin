# 远程终端插件 - 黑名单配置说明

## 功能简介

远程终端插件支持自定义黑名单功能，可以拦截危险命令的执行。

## 默认黑名单

### Linux 默认黑名单（17条）
- `rm -rf` - 删除系统文件
- `dd` - 磁盘写入操作
- `mkfs` - 格式化文件系统
- `shutdown/reboot/poweroff` - 系统关机/重启
- `kill -9 / killall -9` - 强制终止进程
- `chmod/chown -R /` - 递归修改系统权限
- `useradd/userdel` - 用户管理
- `groupadd/groupdel` - 组管理
- `fdisk/parted` - 磁盘分区操作
- `iptables/ufw` - 防火墙规则修改
- `crontab -e/-r` - 编辑/删除定时任务
- `systemctl/service stop/restart/disable` - 停止系统服务
- `init` - 切换系统运行级别
- `su/sudo` - 权限提升
- `passwd` - 修改密码
- 其他系统级危险操作

### Windows 默认黑名单（16条）
- `del/erase` - 删除系统文件
- `rd/rmdir` - 删除系统目录
- `format` - 格式化磁盘
- `shutdown/restart/logoff` - 系统关机/重启/注销
- `taskkill /f` - 强制终止进程
- `net user/group delete` - 删除用户/组
- `diskpart` - 磁盘分区操作
- `bcdedit` - 启动配置编辑
- `reg delete` - 删除注册表项
- `sc stop/delete/config` - 停止/删除系统服务
- `icacls` - 修改系统权限
- `move/xcopy` - 移动/复制系统文件
- 其他系统级危险操作

## 自定义黑名单

### 配置文件位置
```
data/remoteterminal/blacklist.txt
```

### 配置文件格式
每行一条规则，格式为：
```
正则表达式|原因说明
```

支持使用 `#` 开头的注释行

### 配置示例
```txt
# 远程终端插件 - 自定义黑名单配置文件
# 每行一条规则，格式: 正则表达式|原因说明
# 支持使用 # 开头的注释行

# 阻止 Git 硬重置
^git\s+reset\s+--hard|Git 硬重置可能丢失数据

# 阻止强制删除 Docker 容器
^docker\s+(rm|stop)\s+-f|强制删除或停止容器

# 阻止删除 Node.js 包
^npm\s+(uninstall|prune)|删除 Node.js 包

# 阻止卸载 Python 包
^pip\s+uninstall|卸载 Python 包

# 阻止删除系统软件包
^apt\s+(remove|purge)|删除 Debian/Ubuntu 软件包
^yum\s+remove|删除 RedHat/CentOS 软件包
```

### 重新加载配置
修改配置文件后，使用以下命令重新加载：
```
/terminal reload
```

### 查看当前黑名单
```
/terminal list_blacklist
```

## 使用示例

### Linux 示例
```bash
# 允许的命令
/terminal exec ls -la
/terminal exec docker ps
/terminal exec python --version
/terminal exec tail -f /var/log/syslog

# 被拦截的命令（取决于黑名单配置）
/terminal exec rm -rf /          # ❌ 拦截: 删除系统文件
/terminal exec shutdown -h now   # ❌ 拦截: 系统关机
/terminal exec chmod -R 777 /   # ❌ 拦截: 递归修改系统权限
```

### Windows 示例
```bash
# 允许的命令
/terminal exec dir
/terminal exec ipconfig
/terminal exec tasklist
/terminal exec netstat -an

# 被拦截的命令（取决于黑名单配置）
/terminal exec del C:\Windows\*     # ❌ 拦截: 删除系统文件
/terminal exec shutdown -s -t 0     # ❌ 拦截: 系统关机
```

## 新增命令

### /terminal reload
重新加载黑名单配置文件

### /terminal list_blacklist
列出当前所有黑名单规则（包括默认和自定义）

## 注意事项

1. 正则表达式区分大小写（Windows 命令使用 `(?i)` 标志实现不区分大小写）
2. 自定义规则会追加到默认黑名单之后
3. 修改配置文件后需要使用 `reload` 命令重新加载
4. 正则表达式需要符合 Go 语言的正则语法
5. 系统会自动根据运行环境（Windows/Linux）加载对应的默认黑名单
