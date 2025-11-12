[根目录](../../CLAUDE.md) > [pkg](../) > **config**

# config - 配置管理模块

> 更新时间：2025-11-12 15:30:00

## 模块职责

config 模块负责 WatchDucker 的配置管理，包括：
- 命令行参数解析和处理
- 环境变量配置读取
- 配置文件加载（如果需要）
- 全局配置的初始化和访问

## 入口与启动

### 配置初始化
```go
// 在 main 函数入口调用
if err := config.Load(); err != nil {
    logger.Fatal("初始化失败: %v", err)
}
```

### 全局访问
```go
// 在任何模块中访问配置
cfg := config.Get()
if cfg.CheckLabel() {
    // 处理标签模式
}
```

## 对外接口

### 主要导出函数
- `Load() error` - 加载并初始化全局配置
- `Get() *Config` - 获取全局配置实例
- `PrintUsage()` - 打印命令行使用说明

### Config 结构体的 getter 方法
- `CheckLabel() bool` - 是否使用标签模式
- `CheckAll() bool` - 是否检查所有容器
- `NoRestart() bool` - 是否只检查不重启
- `RunOnce() bool` - 是否单次执行模式
- `CronExpression() string` - 获取 cron 表达式
- `ContainerNames() []string` - 获取容器名称列表
- `CleanUp() bool` - 是否在更新后清理悬空镜像

## 关键依赖与配置

### 第三方依赖
- `github.com/spf13/pflag`: 命令行参数解析
- `github.com/spf13/viper`: 配置管理和环境变量

### 配置优先级（由高到低）
1. 命令行参数
2. 环境变量
3. 配置文件
4. 默认值

### 支持的环境变量
- `WATCHDUCKER_LABEL` - 等同于 `--label`
- `WATCHDUCKER_ALL` - 等同于 `--all`
- `WATCHDUCKER_NO_RESTART` - 等同于 `--no-restart`
- `WATCHDUCKER_CRON` - 等同于 `--cron`
- `WATCHDUCKER_CLEAN` - 等同于 `--clean`
- `WATCHDUCKER_LOG_LEVEL` - 设置日志级别

## 数据模型

### Config 结构体
```go
type Config struct {
    checkLabel     bool     // 标签模式标识
    checkAll       bool     // 检查所有容器标识
    noRestart      bool     // 不重启标识
    runOnce        bool     // 单次执行模式
    cronExpression string   // cron 表达式
    containerNames []string // 容器名称列表
    cleanUp        bool     // 清理悬空镜像标识
    logLevel       string   // 日志级别
}
```

### 配置生命周期
1. **初始化阶段**: 解析参数，设置默认值
2. **验证阶段**: 检查配置合法性
3. **使用阶段**: 只读访问，线程安全
4. **清理阶段**: 无需手动清理，随程序结束

## 测试与质量

> ⚠️ **测试状态**: 无单元测试

### 关键测试点
- 命令行参数解析的正确性
- 环境变量的优先级测试
- 配置值的类型转换测试
- 错误处理（无效参数、缺失参数等）

### 测试建议
- 使用 Go 的 flag 测试框架
- 模拟不同的参数组合
- 边界值和异常值测试

## 常见问题 (FAQ)

### Q: 配置加载在哪里调用最合适？
A: 在 `main` 函数最开始调用，确保所有模块都能访问正确的配置

### Q: 如何添加新的配置项？
A: 在 Config 结构体中添加字段，并添加相应的 getter 方法

### Q: 配置是线程安全的吗？
A: 是的，配置在初始化后是只读的，全局访问是线程安全的

### Q: 支持配置文件吗？
A: 当前主要支持命令行和环境变量，可以根据需要扩展配置文件支持

## 相关文件清单

- `config.go` - 配置管理的主要实现

## 变更记录 (Changelog)

### 2025-11-12 15:30:00
- 添加 `--clean` 参数支持，用于在容器更新后清理悬空镜像
- 添加 `CleanUp()` getter 方法
- 添加 `WATCHDUCKER_CLEAN` 环境变量支持
- 更新配置结构体添加 `cleanUp` 字段

### 2025-11-11 14:11:43
- 初始化模块文档
- 记录配置系统架构
- 梳理参数和环境变量映射
- 识别测试需求