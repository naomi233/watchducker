[根目录](../../CLAUDE.md) > [internal](../) > **core**

# core - 核心业务逻辑模块

> 更新时间：2025-11-12 15:30:00

## 模块职责

core 模块承载了 WatchDucker 最核心的业务逻辑，包括：
- **镜像检查器（Checker）**: 检查容器使用的镜像是否有更新版本
- **容器操作器（Operator）**: 执行容器的停止、删除、创建、启动等更新操作
- **批量检查与更新协调**: 并发处理多个容器的检查和更新流程

## 入口与启动

### Checker 创建
```go
checker, err := core.NewChecker()
if err != nil {
    logger.Fatal("创建检查器失败: %v", err)
}
defer checker.Close()
```

### Operator 创建
```go
operator, err := core.NewOperator()
if err != nil {
    logger.Fatal("创建操作器失败: %v", err)
}
defer operator.Close()
```

## 对外接口

### Checker 主要方法
- `CheckByName(ctx, containerNames)` - 按容器名称检查更新
- `CheckByLabel(ctx, labelKey, labelValue)` - 按标签检查更新
- `CheckAll(ctx)` - 检查所有容器
- `Close()` - 清理资源

### Operator 主要方法
- `UpdateContainersByBatchCheckResult(ctx, result)` - 根据检查结果更新容器
- `CleanDanglingImages(ctx)` - 清理悬空镜像
- `Close()` - 清理资源

## 关键依赖与配置

### 依赖模块
- `internal/docker`: Docker API 服务封装
- `internal/types`: 数据类型定义
- `pkg/logger`: 日志记录
- `pkg/utils`: 并发处理和显示工具

### 核心业务流程
1. **容器发现**: 通过 Docker API 获取容器信息
2. **镜像检查**: 并发检查镜像仓库获取更新信息
3. **结果汇总**: 收集检查结果和统计信息
4. **容器更新**: 停止旧容器 → 使用新镜像创建新容器 → 启动新容器
5. **镜像清理**: 在启用 `--clean` 参数时，清理悬空镜像

## 数据模型

### 主要数据结构
- `BatchCheckResult`: 批量检查结果
- `ContainerInfo`: 容器基础信息
- `ImageCheckResult`: 镜像检查结果
- `CheckMode`: 检查模式枚举

## 测试与质量

> ⚠️ **测试状态**: 无单元测试

### 关键测试点
- 镜像检查的并发安全性
- 容器更新操作的原子性
- 错误处理和数据一致性
- Docker API 调用的异常处理

### 测试建议
- 使用 Mock Docker 客户端进行单元测试
- 集成测试需要实际的 Docker 环境
- 异常场景测试（网络中断、镜像不存在等）

## 常见问题 (FAQ)

### Q: 检查器为什么需要 Close 方法？
A: 用于清理底层的 Docker 客户端连接，避免资源泄漏

### Q: 如何处理镜像检查的网络超时？
A: 需要在 Checker 中实现超时控制和重试机制

### Q: 容器更新失败如何回滚？
A: 当前未实现回滚机制，建议在生产环境前充分测试

## 相关文件清单

- `checker.go` - 镜像检查器实现
- `operator.go` - 容器操作器实现

## 变更记录 (Changelog)

### 2025-11-12 15:30:00
- 扩展 Operator 结构体，添加 `imageSvc` 字段
- 添加 `CleanDanglingImages()` 方法，支持悬空镜像清理
- 更新 `NewOperator()` 函数，初始化 ImageService

### 2025-11-11 14:11:43
- 初始化模块文档
- 梳理核心业务流程
- 记录接口和依赖关系
- 识别测试缺口