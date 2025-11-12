[根目录](../CLAUDE.md) > **cmd**

# cmd - 命令行入口与调度模块

> 更新时间：2025-11-12 15:30:00

## 模块职责

cmd 模块是 WatchDucker 的命令行入口和任务调度中心，负责：
- 命令行参数解析和执行模式的确定
- 单次执行模式的处理
- 定时调度器的创建和管理
- 检查组件的协调调用

## 入口与启动

### 主入口流程
```go
main.go → config.Load() → cmd.RunOnce() / cmd.RunCronScheduler()
```

### 执行模式
1. **单次模式**：`--once` 参数触发，执行一次检查后退出
2. **定时模式**：`--cron` 参数触发，持续运行调度器

## 对外接口

### 导出函数
- `RunOnce(ctx context.Context)` - 单次执行入口
- `RunCronScheduler(ctx context.Context)` - 定时调度器入口
- `RunChecker(ctx context.Context, checkFunc func(*core.Checker))` - 通用检查器执行函数

### 配置参数
- `--label`: 检查带有 `watchducker.update=true` 标签的容器
- `--all`: 检查所有容器，无论是否带有标签
- `--no-restart`: 只检查不重启容器
- `--cron`: 定时任务表达式
- `--once`: 单次执行模式
- `--clean`: 更新容器后自动清理悬空镜像

## 关键依赖与配置

### 依赖模块
- `pkg/config`: 获取全局配置
- `pkg/logger`: 日志记录
- `pkg/utils`: 显示工具
- `internal/core`: 核心业务逻辑
- `internal/types`: 数据类型定义

### 第三方依赖
- `github.com/robfig/cron/v3`: 定时任务调度
- `github.com/spf13/pflag`: 命令行参数解析
- `github.com/spf13/viper`: 配置管理

## 数据模型

### 核心数据结构
- `BatchCheckResult`: 批量检查结果汇总
- `ContainerInfo`: 容器基本信息
- `ImageCheckResult`: 镜像检查结果

## 测试与质量

> ⚠️ **测试状态**: 无单元测试

### 测试建议
- 命令行参数解析测试
- 调度器逻辑测试
- 错误处理测试
- 并发安全测试

## 常见问题 (FAQ)

### Q: 如何调试调度器执行？
A: 设置环境变量 `WATCHDUCKER_LOG_LEVEL=DEBUG` 查看详细日志

### Q: 定时任务不执行怎么办？
A: 检查 cron 表达式是否正确，确保程序持续运行且未退出

### Q: 如何手动触发一次检查？
A: 使用 `--once` 参数，如：`watchducker --once --label`

## 相关文件清单

- `cmd.go` - 主要的命令逻辑实现
- `main.go` - 程序入口点（在根目录）

## 变更记录 (Changelog)

### 2025-11-12 15:30:00
- 在 `RunChecker` 函数中集成镜像清理逻辑
- 添加 `--clean` 参数处理，在容器更新成功后执行镜像清理
- 添加适当的日志记录和错误处理

### 2025-11-11 14:11:43
- 初始化模块文档
- 记录接口和依赖关系
- 识别测试缺口