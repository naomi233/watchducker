package config

import (
	"fmt"

	"watchducker/pkg/logger"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	checkLabel     bool     `mapstructure:"label"`
	checkAll       bool     `mapstructure:"all"`
	runOnce        bool     `mapstructure:"-"`
	cronExpression string   `mapstructure:"cron"`
	containerNames []string `mapstructure:"-"` // 位置参数，不通过mapstructure绑定
	noRestart      bool     `mapstructure:"no_restart"`
	cleanUp        bool     `mapstructure:"clean_up"`
	logLevel       string   `mapstructure:"log_level"`
	includeStopped bool     `mapstructure:"include_stopped"`
}

// 全局配置实例（只读，初始化后不可修改）
var globalConfig *Config

// Load 加载配置并初始化全局配置实例
func Load() error {
	// 如果已经初始化，直接返回
	if globalConfig != nil {
		return nil
	}

	config, err := loadConfig()
	if err != nil {
		return err
	}

	globalConfig = config
	return nil
}

// Get 获取全局配置实例（只读访问）
func Get() *Config {
	return globalConfig
}

// CheckLabel 获取 CheckLabel 配置
func (c *Config) CheckLabel() bool {
	return c.checkLabel
}

// CheckAll 获取 CheckAll 配置
func (c *Config) CheckAll() bool {
	return c.checkAll
}

// NoRestart 获取 NoRestart 配置
func (c *Config) NoRestart() bool {
	return c.noRestart
}

// RunOnce 获取 RunOnce 配置
func (c *Config) RunOnce() bool {
	return c.runOnce
}

// CronExpression 获取 CronExpression 配置
func (c *Config) CronExpression() string {
	return c.cronExpression
}

// ContainerNames 获取 ContainerNames 配置
func (c *Config) ContainerNames() []string {
	return c.containerNames
}

// LogLevel 获取 LogLevel 配置
func (c *Config) LogLevel() string {
	return c.logLevel
}

// CleanUp 获取 CleanUp 配置
func (c *Config) CleanUp() bool {
	return c.cleanUp
}

// IncludeStopped 获取 IncludeStopped 配置
func (c *Config) IncludeStopped() bool {
	return c.includeStopped
}

// loadConfig 执行实际的配置加载逻辑
func loadConfig() (*Config, error) {
	// 创建 Viper 实例
	v := viper.New()
	v.SetEnvPrefix("WATCHDUCKER")
	v.AutomaticEnv()

	// 设置 Viper 默认值
	v.SetDefault("label", false)
	v.SetDefault("all", false)
	v.SetDefault("no-restart", false)
	v.SetDefault("cron", "0 2 * * *")
	v.SetDefault("clean", false)
	v.SetDefault("include-stopped", false)

	// 设置命令行参数
	pflag.Bool("label", false, "检查所有带有 watchducker.update=true 标签的容器")
	pflag.Bool("all", false, "检查所有容器，无论是否带有标签")
	pflag.Bool("no-restart", false, "只更新镜像，不重启容器")
	pflag.Bool("once", false, "只执行一次检查和更新，然后退出")
	pflag.String("cron", "0 2 * * *", "定时执行，使用标准 cron 表达式格式")
	pflag.Bool("clean", false, "更新容器后自动清理悬空镜像")
	pflag.Bool("include-stopped", false, "检查时包含已停止的容器")

	// 解析命令行参数
	pflag.Parse()

	// 绑定命令行参数到 Viper
	v.BindPFlags(pflag.CommandLine)

	// 创建配置实例
	config := &Config{
		checkLabel:     v.GetBool("label"),
		checkAll:       v.GetBool("all"),
		noRestart:      v.GetBool("no-restart"),
		runOnce:        v.GetBool("once"),
		cronExpression: v.GetString("cron"),
		// 获取位置参数（容器名称）
		containerNames: pflag.Args(),
		cleanUp:        v.GetBool("clean"),
		logLevel:       v.GetString("LOG_LEVEL"),
		includeStopped: v.GetBool("include-stopped"),
	}

	// 设置日志级别
	if config.logLevel != "" {
		logger.SetLevel(config.logLevel)
	}

	// 验证配置有效性
	if err := config.validate(); err != nil {
		PrintUsage()
		return nil, err
	}

	return config, nil
}

// Validate 验证配置的有效性
func (c *Config) validate() error {
	// 验证至少需要一种检查方式
	if len(c.containerNames) == 0 && !c.checkLabel && !c.checkAll {
		return fmt.Errorf("必须指定容器名称或使用 --label 或 --all")
	}

	return nil
}

// PrintUsage 打印使用方法
func PrintUsage() {
	fmt.Println("\n使用方法:")
	fmt.Println("  watchducker [选项] [容器名称...]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --label       检查所有带有 watchducker.update=true 标签的容器")
	fmt.Println("  --all         检查所有容器，无论是否带有标签")
	fmt.Println("  --no-restart  只更新镜像，不重启容器")
	fmt.Println("  --cron        定时执行，使用标准 cron 表达式格式，默认为 \"0 2 * * *\"")
	fmt.Println("  --once        只执行一次检查和更新，然后退出")
	fmt.Println("  --clean       更新容器后自动清理悬空镜像")
	fmt.Println("  --include-stopped 检查时包含已停止的容器（默认仅检查运行中容器）")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  WATCHDUCKER_LABEL        等同于 --label 选项")
	fmt.Println("  WATCHDUCKER_ALL          等同于 --all 选项")
	fmt.Println("  WATCHDUCKER_NO_RESTART   等同于 --no-restart 选项")
	fmt.Println("  WATCHDUCKER_CRON         等同于 --cron 选项，默认为 0 2 * * *")
	fmt.Println("  WATCHDUCKER_CLEAN        等同于 --clean 选项")
	fmt.Println("  WATCHDUCKER_INCLUDE_STOPPED  等同于 --include-stopped 选项")
	fmt.Println("  WATCHDUCKER_LOG_LEVEL    设置日志级别 (DEBUG/INFO/WARN/ERROR)")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  要检查的容器名称列表（支持多个）  <容器1> <容器2> ... ")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 检查指定容器")
	fmt.Println("  watchducker --once nginx redis mysql")
	fmt.Println()
	fmt.Println("  # 检查所有带有 watchducker.update=true 标签的容器")
	fmt.Println("  watchducker --label --once")
	fmt.Println()
	fmt.Println("  # 检查所有容器")
	fmt.Println("  watchducker --all --once")
	fmt.Println()
	fmt.Println("  # 使用环境变量配置")
	fmt.Println("  export WATCHDUCKER_LOG_LEVEL=DEBUG")
	fmt.Println("  export WATCHDUCKER_LABEL=true")
	fmt.Println("  export WATCHDUCKER_ALL=true")
	fmt.Println("  export WATCHDUCKER_CRON=\"0 2 * * *\"")
	fmt.Println()
	fmt.Println("  # 定时执行示例")
	fmt.Println("  watchducker --cron \"0 2 * * *\" --label          # 每天凌晨2点检查所有标签容器")
	fmt.Println("  watchducker --cron \"*/30 * * * *\" nginx redis   # 每30分钟检查指定容器")
	fmt.Println("  watchducker --cron \"@daily\" --all               # 每天检查所有容器")
	fmt.Println("  watchducker --cron \"@daily\" --no-restart        # 每天执行，只检查不重启")
	fmt.Println("  watchducker --cron \"@daily\" --clean             # 每天执行并清理悬空镜像")
	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  - 优先级：指定容器 > --label > --all")
	fmt.Println("  - 环境变量优先级低于命令行参数")
}
