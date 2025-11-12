# WatchDucker 🐤🦆

一个用 Go 语言编写的 Docker 容器镜像更新检查和自动更新工具。

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

## ✨ 特性

- 🔍 **智能检查**: 自动检测容器使用的镜像是否有新版本可用
- 🏷️ **标签驱动**: 通过 `watchducker.update=true` 标签自动管理需要更新的容器
- ⏰ **定时执行**: 支持使用 cron 表达式进行定时检查
- 🔄 **自动更新**: 检测到更新后可自动重启容器使用新镜像
- 🚫 **灵活控制**: 提供只检查不重启的选项
- ✨ **实时反馈**: 检查过程中提供实时进度和结果输出
- 🐳 **Docker 原生**: 完全基于 Docker API，无需额外依赖
- ⚙️ **无需代理**: 复用现有 Docker 配置，无需额外配置认证和代理、[加速镜像源](https://github.com/dongyubin/DockerHub)

## 🚀 快速开始

### 二进制安装

从 [Releases 页面](https://github.com/naomi233/watchducker/releases) 下载对应平台的二进制文件：

### Docker 镜像

```bash
docker pull naomi233/watchducker:latest
```

### 源码编译

```bash
git clone https://github.com/naomi233/watchducker.git
cd watchducker
go build -o watchducker .
```

## 📖 使用方法

### Docker

```bash
# 检查指定容器一次
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --once nginx redis mysql
# 检查所有带有更新标签的容器一次
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --label --once
# 检查所有容器一次，更新后清理悬空镜像
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --all --clean --once
# 只更新镜像，不重启容器
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --no-restart --once nginx redis
# 使用标签模式，同时防止自动重启
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --label --no-restart --once
# 每天凌晨2点检查所有标签容器
docker run --name watchducker -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --cron "0 2 * * *" --label
# 每30分钟检查指定容器
docker run --name watchducker -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --cron "*/30 * * * *" nginx redis
# 每天执行，只检查不重启
docker run --name watchducker -v /var/run/docker.sock:/var/run/docker.sock naomi233/watchducker:latest watchducker --cron "@daily" --no-restart nginx
```

### 可执行文件

```bash
# 检查指定容器一次
watchducker --once nginx redis mysql
# 检查所有带有更新标签的容器一次
watchducker --label --once
# 检查所有容器一次，更新后清理悬空镜像
watchducker --all --clean --once
# 只更新镜像，不重启容器
watchducker --no-restart --once nginx redis
# 使用标签模式，同时防止自动重启
watchducker --label --no-restart --once
# 每天凌晨2点检查所有标签容器
watchducker --cron "0 2 * * *" --label
# 每30分钟检查指定容器
watchducker --cron "*/30 * * * *" nginx redis
# 每天执行，只检查不重启
watchducker --cron "@daily" --no-restart nginx
```

### Docker Compose 配置示例

```yml
services:
  watchducker:
    image: naomi233/watchducker
    container_name: watchducker
    network_mode: bridge
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - TZ=Asia/Shanghai
      - WATCHDUCKER_LOG_LEVEL=DEBUG
      - WATCHDUCKER_CRON=0 2 * * *
      - WATCHDUCKER_LABEL=true
```

## ⚙️ 配置选项

### 命令行参数

- `--label`: 检查所有带有 `watchducker.update=true` 标签的容器
- `--no-restart`: 只更新镜像，不重启容器
- `--all`: 检查所有容器
- `--clean`: 更新容器后自动清理悬空镜像
- `--cron`: 定时执行，使用标准 [cron 表达式](https://crontab.guru) 格式，默认 "0 2 * * *"
- `--once`: 只执行一次检查和更新，然后退出
- 容器名称列表

### 环境变量

```bash
# 等同于 --label 选项
export WATCHDUCKER_LABEL=true

# 等同于 --label 选项
export WATCHDUCKER_ALL=true

# 等同于 --no-restart 选项
export WATCHDUCKER_NO_RESTART=true

# 等同于 --clean 选项
export WATCHDUCKER_CLEAN=true

# 等同于 --cron 选项
export WATCHDUCKER_CRON="0 2 * * *"

# 设置日志级别 (DEBUG/INFO/WARN/ERROR)
export WATCHDUCKER_LOG_LEVEL=DEBUG
```

### 使用标签驱动更新

为需要自动更新的容器添加标签：

```bash
docker run --name nginx --label watchducker.update=true nginx:latest
```

## 🏗️ 项目架构

### 目录结构

```
watchducker/
├── cmd/                    # 命令行入口
│   └── cmd.go               # 主命令逻辑
├── internal/                 # 内部模块
│   ├── core/                # 核心业务逻辑
│   │   ├── checker.go         # 镜像检查器
│   │   └── operator.go      # 容器操作器
│   ├── docker/               # Docker API 封装
│   │   ├── client.go         # 客户端管理
│   │   ├── container.go     # 容器服务
│   │   └── image.go          # 镜像服务
│   └── types/                 # 类型定义
│       └── types.go
├── pkg/                      # 可复用的公共包
│   ├── config/                # 配置管理
│   │   └── config.go
├── pkg/logger/               # 日志系统
│   │   └── logger.go
│   └── utils/                 # 工具函数
│       └── display.go         # 显示输出
├── main.go                    # 程序入口
```

### 核心组件

- **Checker**: 镜像检查器，负责检查容器使用的镜像是否有更新
- **Operator**: 容器操作器，负责容器的重启和更新操作
- **ContainerService**: 容器服务，封装 Docker 容器的操作
- **ImageService**: 镜像服务，封装 Docker 镜像的检查逻辑

## 🔧 开发

### 依赖要求

- Go 1.25 或更高版本
- Docker 守护进程（用于容器操作）
- 网络连接（用于镜像仓库访问）

### 项目构建

```bash
# 开发构建
go build -o watchducker .

# 多平台发布（使用 GoReleaser）
goreleaser build --snapshot

# 创建 Docker 镜像
docker build -t watchducker .
```

## 📊 工作流程

1. **容器发现**: 根据容器名称或标签查找相关容器
2. **镜像检查**: 并发检查所有镜像是否有更新版本
3. **自动更新**: 停止旧容器 → 删除旧容器 → 创建新容器 → 启动新容器

## 🔐 安全性

- 只对指定标签的容器进行操作
- 提供清晰的日志记录所有操作
- 支持只检查模式，避免意外重启

## 🐛 故障排除

### 常见问题

1. **权限错误**: 确保程序有足够的权限访问 Docker 守护进程
2. **网络连接**: 检查是否有网络连接访问镜像仓库
3. **容器状态**: 确保目标容器处于运行状态

### 调试模式

```bash
# 启用调试日志
export WATCHDUCKER_LOG_LEVEL=DEBUG
watchducker --label
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 [GNU GPL v3](LICENSE) 许可证。

---

**WatchDucker** - 让 Docker 容器更新变得简单智能！

> ⚠️ **注意**: 在生产环境中使用前，请充分测试所有功能。
