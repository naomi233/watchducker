package config

import (
	"fmt"
	"strings"

	"watchducker/pkg/logger"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	logLevel           string   `mapstructure:"log_level"`
	containerNames     []string `mapstructure:"-"` // 位置参数，不通过mapstructure绑定
	checkAll           bool     `mapstructure:"all"`
	checkLabel         bool     `mapstructure:"label"`
	checkLabelReversed bool     `mapstructure:"label_reversed"`
	cronExpression     string   `mapstructure:"cron"`
	runOnce            bool     `mapstructure:"-"`
	cleanUp            bool     `mapstructure:"clean_up"`
	noRestart          bool     `mapstructure:"no_restart"`
	includeStopped     bool     `mapstructure:"include_stopped"`
	disabledContainers string   `mapstructure:"disabled_containers"`
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

// LogLevel 获取 LogLevel 配置
func (c *Config) LogLevel() string {
	return c.logLevel
}

// ContainerNames 获取 ContainerNames 配置
func (c *Config) ContainerNames() []string {
	return c.containerNames
}

// CheckAll 获取 CheckAll 配置
func (c *Config) CheckAll() bool {
	return c.checkAll
}

// CheckLabel 获取 CheckLabel 配置
func (c *Config) CheckLabel() bool {
	return c.checkLabel
}

// CheckLabelReversed 获取 CheckLabelReversed 配置
func (c *Config) CheckLabelReversed() bool {
	return c.checkLabelReversed
}

// CronExpression 获取 CronExpression 配置
func (c *Config) CronExpression() string {
	return c.cronExpression
}

// RunOnce 获取 RunOnce 配置
func (c *Config) RunOnce() bool {
	return c.runOnce
}

// CleanUp 获取 CleanUp 配置
func (c *Config) CleanUp() bool {
	return c.cleanUp
}

// NoRestart 获取 NoRestart 配置
func (c *Config) NoRestart() bool {
	return c.noRestart
}

// IncludeStopped 获取 IncludeStopped 配置
func (c *Config) IncludeStopped() bool {
	return c.includeStopped
}

// DisabledContainers 获取被排除的容器列表
func (c *Config) DisabledContainers() []string {
	return strings.Split(c.disabledContainers, ",")
}

// loadConfig 执行实际的配置加载逻辑
func loadConfig() (*Config, error) {
	// 创建 Viper 实例
	v := viper.New()
	v.SetEnvPrefix("WATCHDUCKER")
	v.AutomaticEnv()

	// 设置 Viper 默认值
	v.SetDefault("all", false)
	v.SetDefault("label", false)
	v.SetDefault("label-reversed", false)
	v.SetDefault("cron", "0 2 * * *")
	v.SetDefault("clean", false)
	v.SetDefault("no-restart", false)
	v.SetDefault("include-stopped", false)
	v.SetDefault("disabled-containers", "")

	// 环境变量键名中的连字符替换为下划线
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// 设置命令行参数
	pflag.Bool("all", false, "检查所有容器，无论是否带有标签")
	pflag.Bool("label", false, "检查带有 watchducker.update=true 标签的容器")
	pflag.Bool("label-reversed", false, "检查没有 watchducker.update=true 标签的容器")
	pflag.String("cron", "0 2 * * *", "定时执行，使用标准 cron 表达式格式")
	pflag.Bool("once", false, "只执行一次检查和更新，然后退出")
	pflag.Bool("clean", false, "更新容器后自动清理悬空镜像")
	pflag.Bool("no-restart", false, "只更新镜像，不重启容器")
	pflag.Bool("include-stopped", false, "检查时包含已停止的容器")
	pflag.String("disabled-containers", "", "排除指定的容器，不进行检查和更新")

	// 解析命令行参数
	pflag.Parse()

	// 绑定命令行参数到 Viper
	v.BindPFlags(pflag.CommandLine)

	config := &Config{
		containerNames:     pflag.Args(), // 获取位置参数（容器名称）
		logLevel:           v.GetString("LOG_LEVEL"),
		checkAll:           v.GetBool("all"),
		checkLabel:         v.GetBool("label"),
		checkLabelReversed: v.GetBool("label-reversed"),
		noRestart:          v.GetBool("no-restart"),
		runOnce:            v.GetBool("once"),
		cronExpression:     v.GetString("cron"),
		cleanUp:            v.GetBool("clean"),
		includeStopped:     v.GetBool("include-stopped"),
		disabledContainers: v.GetString("disabled-containers"),
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
	if len(c.containerNames) == 0 && !c.checkLabel && !c.checkAll && !c.checkLabelReversed {
		return fmt.Errorf("必须指定容器名称或使用 --label 或 --all 或 --label-reversed 选项")
	}

	return nil
}

// PrintUsage 打印使用方法
func PrintUsage() {
	fmt.Println("\n使用方法:")
	fmt.Println("  watchducker [选项] [容器名称...]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --all                 检查所有容器，无论是否带有标签")
	fmt.Println("  --label               检查带有 watchducker.update=true 标签的容器")
	fmt.Println("  --label-reversed      检查没有 watchducker.update=true 标签的容器")
	fmt.Println("  --cron                定时执行，使用标准 cron 表达式格式，默认为 \"0 2 * * *\"")
	fmt.Println("  --once                只执行一次检查和更新，然后退出")
	fmt.Println("  --clean               更新容器后自动清理悬空镜像")
	fmt.Println("  --no-restart          只更新镜像，不重启容器")
	fmt.Println("  --include-stopped     检查时包含已停止的容器（默认仅检查运行中容器）")
	fmt.Println("  --disabled-containers 排除指定的容器，不进行检查和更新")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  WATCHDUCKER_LOG_LEVEL           设置日志级别 (DEBUG/INFO/WARN/ERROR)")
	fmt.Println("  WATCHDUCKER_ALL                 等同于 --all 选项")
	fmt.Println("  WATCHDUCKER_LABEL               等同于 --label 选项")
	fmt.Println("  WATCHDUCKER_LABEL_REVERSED      等同于 --label-reversed 选项")
	fmt.Println("  WATCHDUCKER_CRON                等同于 --cron 选项，默认为 0 2 * * *")
	fmt.Println("  WATCHDUCKER_CLEAN               等同于 --clean 选项")
	fmt.Println("  WATCHDUCKER_NO_RESTART          等同于 --no-restart 选项")
	fmt.Println("  WATCHDUCKER_INCLUDE_STOPPED     等同于 --include-stopped 选项")
	fmt.Println("  WATCHDUCKER_DISABLED_CONTAINERS 等同于 --disabled-containers 选项")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  要检查的容器名称列表（支持多个）  <容器1> <容器2> ... ")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 检查指定容器")
	fmt.Println("  watchducker --once nginx redis mysql")
	fmt.Println()
	fmt.Println("  # 检查所有容器")
	fmt.Println("  watchducker --all --once")
	fmt.Println()
	fmt.Println("  # 检查带有 watchducker.update=true 标签的容器")
	fmt.Println("  watchducker --label --once")
	fmt.Println()
	fmt.Println("  # 检查没有 watchducker.update=true 标签的容器")
	fmt.Println("  watchducker --label-reversed --once")
	fmt.Println()
	fmt.Println("  # 定时执行示例")
	fmt.Println("  watchducker --cron \"0 2 * * *\" --label --clean                # 每天凌晨2点检查更新所有标签容器，清理悬空镜像")
	fmt.Println("  watchducker --cron \"*/30 * * * *\" nginx redis                 # 每30分钟检查更新指定nginx、redis容器")
	fmt.Println("  watchducker --cron \"@daily\" --all --disabled-containers mysql # 每天检查更新所有容器，但排除mysql")
	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  - 优先级：指定容器 > --all > --label-reversed > --label")
}
