package utils

import (
	"fmt"
	"time"

	"watchducker/internal/types"
	"watchducker/pkg/logger"
)

// PrintContainerList 打印容器列表
func PrintContainerList(containers []types.ContainerInfo) {
	fmt.Println("\n=== 容器列表 ===")
	if len(containers) == 0 {
		fmt.Println("未找到匹配的容器")
		return
	}

	fmt.Printf("%-12s %-20s %-20s %s\n", "ID", "名称", "镜像", "状态")
	fmt.Println("----------------------------------------------------------------")

	for _, container := range containers {
		fmt.Printf("%-12s %-20s %-20s %s\n",
			container.ID,
			container.Name,
			container.Image,
			container.State)
	}
}

// PrintBatchSummary 打印批量检查的统计信息
func PrintBatchSummary(result *types.BatchCheckResult) {
	fmt.Println("\n=== 统计信息 ===")
	fmt.Printf("匹配的容器数: %d\n", result.Summary.TotalContainers)
	fmt.Printf("检查的镜像数: %d\n", result.Summary.TotalImages)
	fmt.Printf("有更新的镜像: %d\n", result.Summary.Updated)
	fmt.Printf("最新的镜像: %d\n", result.Summary.UpToDate)
	fmt.Printf("检查失败的镜像: %d\n", result.Summary.Failed)
	fmt.Printf("检查耗时: %v\n", result.Summary.Duration.Round(time.Millisecond))
}

// CreateCheckCallback 创建镜像检查回调函数
func CreateCheckCallback() types.CheckCallback {
	return func(info *types.ImageCheckResult) {
		status := "✅ 最新"
		if info.Error != "" {
			status = "❌ 失败"
		} else if info.IsUpdated {
			status = "🔄 有更新"
		}
		logger.Info("镜像 %-20s %s", info.Name, status)
	}
}

func GetUpdateSummary(result *types.BatchCheckResult) string {
	var summary string
	summary += "\n=== 更新信息 ===\n"
	for _, item := range result.Images {
		if item.IsUpdated && item.Error == "" {
			summary += fmt.Sprintf("镜像 %-20s 更新成功✅\n", item.Name)
		} else if item.Error != "" {
			summary += fmt.Sprintf("镜像 %-20s 更新失败❌: %s\n", item.Name, item.Error)
		}
	}
	return summary
}

// PrintWelcome 打印欢迎信息
func PrintWelcome() {
	fmt.Println("========================================")
	fmt.Println("      WatchDucker - Docker 镜像更新检查器")
	fmt.Println("========================================")
}
