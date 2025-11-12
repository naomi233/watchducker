package core

import (
	"context"
	"fmt"
	"time"

	"watchducker/internal/docker"
	"watchducker/internal/types"
	"watchducker/pkg/logger"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// Operator 容器自动更新器
type Operator struct {
	clientManager   *docker.ClientManager
	containerSvc    *docker.ContainerService
	containerOpsSvc *docker.ContainerService
	imageSvc        *docker.ImageService
}

// NewOperator 创建新的更新器实例
func NewOperator() (*Operator, error) {
	clientManager, err := docker.NewClientManager()
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端管理器失败: %w", err)
	}

	containerSvc := docker.NewContainerService(clientManager)
	containerOpsSvc := docker.NewContainerService(clientManager)
	imageSvc := docker.NewImageService(clientManager)

	return &Operator{
		clientManager:   clientManager,
		containerSvc:    containerSvc,
		containerOpsSvc: containerOpsSvc,
		imageSvc:        imageSvc,
	}, nil
}

// createNewContainer 使用新镜像创建新容器
func (u *Operator) createNewContainer(ctx context.Context, containerConfig *dockerTypes.ContainerJSON, newImage string, containerName string) (string, error) {
	// 创建容器配置
	config := &container.Config{
		Image:  newImage,
		Cmd:    containerConfig.Config.Cmd,
		Env:    containerConfig.Config.Env,
		Labels: containerConfig.Config.Labels,
	}

	// 创建主机配置
	hostConfig := &container.HostConfig{
		Binds:         containerConfig.HostConfig.Binds,
		PortBindings:  containerConfig.HostConfig.PortBindings,
		RestartPolicy: containerConfig.HostConfig.RestartPolicy,
		NetworkMode:   containerConfig.HostConfig.NetworkMode,
		VolumesFrom:   containerConfig.HostConfig.VolumesFrom,
	}

	// 创建网络配置
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: containerConfig.NetworkSettings.Networks,
	}

	// 创建新容器
	newContainerID, err := u.containerOpsSvc.CreateContainer(ctx, config, hostConfig, networkingConfig, containerName)
	if err != nil {
		return "", err
	}

	return newContainerID, nil
}

// UpdateContainer 更新容器到新镜像
func (u *Operator) updateContainer(ctx context.Context, containerInfo types.ContainerInfo, newImage string) error {
	logger.Info("开始更新容器 %s (%s) 到新镜像 %s", containerInfo.Name, containerInfo.ID, newImage)

	// 1. 获取容器完整配置
	containerConfig, err := u.containerOpsSvc.GetContainerConfig(ctx, containerInfo.ID)
	if err != nil {
		return fmt.Errorf("获取容器配置失败: %w", err)
	}

	// 2. 停止容器
	stopTimeout := 30 * time.Second
	if err := u.containerOpsSvc.StopContainer(ctx, containerInfo.ID, &stopTimeout); err != nil {
		return fmt.Errorf("停止容器失败: %w", err)
	}

	// 3. 删除容器
	if err := u.containerOpsSvc.RemoveContainer(ctx, containerInfo.ID, true); err != nil {
		return fmt.Errorf("删除容器失败: %w", err)
	}

	// 4. 使用新镜像创建新容器
	newContainerID, err := u.createNewContainer(ctx, containerConfig, newImage, containerInfo.Name)
	if err != nil {
		return fmt.Errorf("创建新容器失败: %w", err)
	}

	// 5. 启动新容器
	if err := u.containerOpsSvc.StartContainer(ctx, newContainerID); err != nil {
		return fmt.Errorf("启动新容器失败: %w", err)
	}

	logger.Info("容器 %s 已成功更新到新镜像 %s，新容器ID: %s", containerInfo.Name, newImage, newContainerID[:12])
	return nil
}

// UpdateContainersWithNewImages 批量更新容器到新镜像
func (u *Operator) updateContainers(ctx context.Context, containers []types.ContainerInfo, imageUpdates map[string]string) error {
	logger.Info("开始批量更新 %d 个容器", len(containers))

	var errors []error

	for _, containerInfo := range containers {
		newImage, exists := imageUpdates[containerInfo.Image]
		if !exists {
			logger.Warn("容器 %s 的镜像 %s 没有找到对应的新镜像，跳过更新", containerInfo.Name, containerInfo.Image)
			continue
		}

		if err := u.updateContainer(ctx, containerInfo, newImage); err != nil {
			logger.Error("更新容器 %s 失败: %v", containerInfo.Name, err)
			errors = append(errors, fmt.Errorf("更新容器 %s 失败: %w", containerInfo.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("批量更新过程中出现 %d 个错误: %v", len(errors), errors)
	}

	logger.Info("批量更新完成，成功更新 %d 个容器", len(containers))
	return nil
}

// UpdateContainers 更新有镜像更新的容器
func (c *Operator) UpdateContainersByBatchCheckResult(ctx context.Context, result *types.BatchCheckResult) error {
	if result.Summary.Updated == 0 {
		logger.Info("没有需要更新的容器")
		return nil
	}

	logger.Info("发现 %d 个容器需要更新，开始自动更新流程", result.Summary.Updated)

	// 构建镜像更新映射
	imageUpdates := make(map[string]string)
	for _, imageResult := range result.Images {
		if imageResult.IsUpdated && imageResult.Error == "" {
			imageUpdates[imageResult.Name] = imageResult.Name // 使用相同的镜像名称，但实际是新版本
		}
	}

	// 更新所有使用这些镜像的容器
	var containersToUpdate []types.ContainerInfo
	for _, container := range result.Containers {
		if _, exists := imageUpdates[container.Image]; exists {
			containersToUpdate = append(containersToUpdate, container)
		}
	}

	if len(containersToUpdate) == 0 {
		logger.Warn("没有找到需要更新的容器")
		return nil
	}

	// 执行批量更新
	if err := c.updateContainers(ctx, containersToUpdate, imageUpdates); err != nil {
		return err
	}

	return nil
}

// CleanDanglingImages 清理悬空镜像
func (u *Operator) CleanDanglingImages(ctx context.Context) error {
	logger.Info("开始清理悬空镜像")

	err := u.imageSvc.CleanDanglingImages(ctx)
	if err != nil {
		return fmt.Errorf("清理悬空镜像失败: %w", err)
	}

	logger.Info("悬空镜像清理完成")
	return nil
}

// Close 关闭所有资源
func (u *Operator) Close() error {
	if u.clientManager != nil {
		return u.clientManager.Close()
	}
	return nil
}
