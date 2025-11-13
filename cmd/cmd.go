package cmd

import (
	"context"

	"watchducker/internal/core"
	"watchducker/internal/types"
	"watchducker/pkg/config"
	"watchducker/pkg/logger"
	"watchducker/pkg/notify"
	"watchducker/pkg/utils"

	"github.com/robfig/cron/v3"
)

// checkContainersByName 根据容器名称检查镜像更新
func checkContainersByName(ctx context.Context) {
	cfg := config.Get()
	RunChecker(ctx, func(checker *core.Checker) (*types.BatchCheckResult, error) {
		return checker.CheckByName(ctx, cfg.ContainerNames())
	})
}

// checkContainersByLabel 根据标签检查镜像更新
func checkContainersByLabel(ctx context.Context) {
	labelKey, labelValue := "watchducker.update", "true"

	RunChecker(ctx, func(checker *core.Checker) (*types.BatchCheckResult, error) {
		return checker.CheckByLabel(ctx, labelKey, labelValue)
	})
}

// checkAllContainers 检查所有容器的镜像更新
func checkAllContainers(ctx context.Context) {
	RunChecker(ctx, func(checker *core.Checker) (*types.BatchCheckResult, error) {
		return checker.CheckAll(ctx)
	})
}

// RunOnce 单次执行模式
func RunOnce(ctx context.Context) {
	cfg := config.Get()

	if len(cfg.ContainerNames()) > 0 {
		checkContainersByName(ctx)
	} else if cfg.CheckLabel() {
		checkContainersByLabel(ctx)
	} else if cfg.CheckAll() {
		checkAllContainers(ctx)
	} else {
		config.PrintUsage()
	}
}

// RunCronScheduler 运行定时调度器
func RunCronScheduler(ctx context.Context) {
	cfg := config.Get()

	// 创建 cron 调度器
	c := cron.New()

	// 添加定时任务
	_, err := c.AddFunc(cfg.CronExpression(), func() {
		logger.Info("定时任务开始执行")

		RunOnce(ctx)

		logger.Info("定时任务执行完成")
	})

	if err != nil {
		logger.Fatal("无效的 cron 表达式 '%s': %v", cfg.CronExpression(), err)
	}

	logger.Info("定时任务已启动，cron 表达式: %s", cfg.CronExpression())
	logger.Info("按 Ctrl+C 停止定时任务")

	// 启动调度器
	c.Start()

	// 保持程序运行
	select {}
}

// RunChecker 创建并运行检查器的通用函数
func RunChecker(ctx context.Context, checkFunc func(*core.Checker) (*types.BatchCheckResult, error)) {
	utils.PrintWelcome()

	cfg := config.Get()

	// 创建检查器
	checker, err := core.NewChecker()
	if err != nil {
		logger.Fatal("创建检查器失败: %v", err)
	}
	defer checker.Close()

	// 使用回调函数实时输出结果
	result, err := checkFunc(checker)
	if err != nil {
		logger.Error("容器检查过程中出现错误: %v", err)
	}

	if result == nil {
		return
	}

	if !cfg.NoRestart() && result.Summary.Updated > 0 {
		// 创建操作器
		operator, err := core.NewOperator()
		if err != nil {
			logger.Fatal("创建操作器失败: %v", err)
		}
		defer operator.Close()

		// 更新有镜像更新的容器
		err = operator.UpdateContainersByBatchCheckResult(ctx, result)
		if err != nil {
			logger.Error("容器更新过程中出现错误: %v", err)
		}

		// 如果启用了清理功能，清理悬空镜像
		if cfg.CleanUp() {
			if err := operator.CleanDanglingImages(ctx); err != nil {
				logger.Error("清理悬空镜像失败: %v", err)
			}
		}

		notify.Send("WatchDucker 镜像更新", utils.GetUpdateSummary(result))
	}

	// 输出最终结果
	utils.PrintContainerList(result.Containers)
	utils.PrintBatchSummary(result)
}
