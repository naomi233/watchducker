package docker

import (
	"context"
	"fmt"
	"time"

	"watchducker/internal/types"
	"watchducker/pkg/logger"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

// ContainerService 容器服务
type ContainerService struct {
	clientManager *ClientManager
}

// NewContainerService 创建容器服务实例
func NewContainerService(clientManager *ClientManager) *ContainerService {
	return &ContainerService{
		clientManager: clientManager,
	}
}

// createContainerInfo 创建容器信息结构体
func (cs *ContainerService) createContainerInfo(container dockerTypes.Container, name string) types.ContainerInfo {
	return types.ContainerInfo{
		ID:     container.ID[:12], // 使用短ID
		Name:   name,
		Image:  container.Image,
		Labels: container.Labels,
		State:  container.State,
	}
}

// GetByName 根据容器名称获取容器信息
func (cs *ContainerService) GetByName(ctx context.Context, containerNames []string, includeStopped bool) ([]types.ContainerInfo, error) {
	cli := cs.clientManager.GetClient()

	// 获取所有容器列表
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: includeStopped,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	var result []types.ContainerInfo
	for _, container := range containers {
		// 检查容器名称是否匹配（支持部分匹配）
		for _, name := range container.Names {
			// 移除开头的斜杠进行匹配
			normalizedName := name
			if len(normalizedName) > 0 && normalizedName[0] == '/' {
				normalizedName = normalizedName[1:]
			}

			for _, containerName := range containerNames {
				if normalizedName == containerName {
					containerInfo := cs.createContainerInfo(container, normalizedName)
					result = append(result, containerInfo)
					break // 找到匹配后跳出内层循环
				}
			}
		}
	}

	return result, nil
}

// GetByLabel 根据标签获取容器信息
func (cs *ContainerService) GetByLabel(ctx context.Context, labelKey, labelValue string, includeStopped bool) ([]types.ContainerInfo, error) {
	cli := cs.clientManager.GetClient()

	filter := filters.NewArgs()

	// 如果指定了标签值，精确匹配；否则只检查标签键存在
	if labelValue != "" {
		filter.Add("label", fmt.Sprintf("%s=%s", labelKey, labelValue))
	} else {
		filter.Add("label", labelKey)
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     includeStopped,
		Filters: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	var result []types.ContainerInfo
	for _, container := range containers {
		// 移除开头的斜杠进行匹配
		normalizedName := container.Names[0]
		if len(normalizedName) > 0 && normalizedName[0] == '/' {
			normalizedName = normalizedName[1:]
		}

		containerInfo := cs.createContainerInfo(container, normalizedName)
		result = append(result, containerInfo)
	}

	return result, nil
}

// StopContainer 停止容器
func (cos *ContainerService) StopContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	cli := cos.clientManager.GetClient()

	logger.Debug("正在停止容器: %s", containerID[:12])

	stopOptions := container.StopOptions{}
	if timeout != nil {
		timeoutSeconds := int(timeout.Seconds())
		stopOptions.Timeout = &timeoutSeconds
	}

	if err := cli.ContainerStop(ctx, containerID, stopOptions); err != nil {
		logger.Error("停止容器 %s 失败: %v", containerID[:12], err)
		return fmt.Errorf("停止容器 %s 失败: %w", containerID[:12], err)
	}

	logger.Debug("容器 %s 已成功停止", containerID[:12])
	return nil
}

// RemoveContainer 删除容器
func (cos *ContainerService) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	cli := cos.clientManager.GetClient()

	logger.Debug("正在删除容器: %s", containerID[:12])

	removeOptions := container.RemoveOptions{
		Force: force,
	}

	if err := cli.ContainerRemove(ctx, containerID, removeOptions); err != nil {
		logger.Error("删除容器 %s 失败: %v", containerID[:12], err)
		return fmt.Errorf("删除容器 %s 失败: %w", containerID[:12], err)
	}

	logger.Debug("容器 %s 已成功删除", containerID[:12])
	return nil
}

// StartContainer 启动容器
func (cos *ContainerService) StartContainer(ctx context.Context, containerID string) error {
	cli := cos.clientManager.GetClient()

	logger.Debug("正在启动容器: %s", containerID[:12])

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		logger.Error("启动容器 %s 失败: %v", containerID[:12], err)
		return fmt.Errorf("启动容器 %s 失败: %w", containerID[:12], err)
	}

	logger.Debug("容器 %s 已成功启动", containerID[:12])
	return nil
}

// CreateContainer 创建容器
func (cos *ContainerService) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error) {
	cli := cos.clientManager.GetClient()

	logger.Debug("正在创建容器: %s", containerName)

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
	if err != nil {
		logger.Error("创建容器 %s 失败: %v", containerName, err)
		return "", fmt.Errorf("创建容器 %s 失败: %w", containerName, err)
	}

	logger.Debug("容器 %s 已成功创建，ID: %s", containerName, resp.ID[:12])
	return resp.ID, nil
}

// GetAll 获取所有容器信息
func (cs *ContainerService) GetAll(ctx context.Context, includeStopped bool) ([]types.ContainerInfo, error) {
	cli := cs.clientManager.GetClient()

	// 获取所有容器列表
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: includeStopped,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	var result []types.ContainerInfo
	for _, container := range containers {
		// 使用第一个名称作为容器名称
		normalizedName := container.Names[0]
		if len(normalizedName) > 0 && normalizedName[0] == '/' {
			normalizedName = normalizedName[1:]
		}

		containerInfo := cs.createContainerInfo(container, normalizedName)
		result = append(result, containerInfo)
	}

	return result, nil
}

// GetContainerConfig 获取容器配置
func (cos *ContainerService) GetContainerConfig(ctx context.Context, containerID string) (*dockerTypes.ContainerJSON, error) {
	cli := cos.clientManager.GetClient()

	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		logger.Error("获取容器 %s 配置失败: %v", containerID[:12], err)
		return nil, fmt.Errorf("获取容器 %s 配置失败: %w", containerID[:12], err)
	}

	return &containerJSON, nil
}
