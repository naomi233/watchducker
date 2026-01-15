package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"watchducker/internal/docker"
	"watchducker/internal/types"
	"watchducker/pkg/logger"
	"watchducker/pkg/utils"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// SelfUpdater WatchDucker 自我更新器
type SelfUpdater struct {
	clientManager   *docker.ClientManager
	containerSvc    *docker.ContainerService
	containerOpsSvc *docker.ContainerService
	imageSvc        *docker.ImageService
}

// NewSelfUpdater 创建新的自我更新器实例
func NewSelfUpdater() (*SelfUpdater, error) {
	clientManager, err := docker.NewClientManager()
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端管理器失败: %w", err)
	}

	containerSvc := docker.NewContainerService(clientManager)
	containerOpsSvc := docker.NewContainerService(clientManager)
	imageSvc := docker.NewImageService(clientManager)

	return &SelfUpdater{
		clientManager:   clientManager,
		containerSvc:    containerSvc,
		containerOpsSvc: containerOpsSvc,
		imageSvc:        imageSvc,
	}, nil
}

// findSelfContainer 定位 WatchDucker 自我容器
func (su *SelfUpdater) findSelfContainer(ctx context.Context) (*types.ContainerInfo, error) {
	// 获取所有容器
	allContainers, err := su.containerSvc.GetAll(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	// 遍历所有容器，找到符合条件的自我容器
	for _, container := range allContainers {
		if su.isSelfContainer(ctx, container) {
			logger.Info("找到 WatchDucker 自我容器: %s (%s)", container.Name, container.ID)
			return &container, nil
		}
	}

	return nil, fmt.Errorf("未找到 WatchDucker 自我容器")
}

// isSelfContainer 检查容器是否是 WatchDucker 自身
func (su *SelfUpdater) isSelfContainer(_ context.Context, containerInfo types.ContainerInfo) bool {
	// 检查容器是否有 WatchDucker 的特定标签
	if containerInfo.Labels != nil && containerInfo.Labels["naomi233.watchducker"] == "true" {
		return true
	}

	// 检查容器镜像是否包含 "watchducker" 关键字
	if containerInfo.Image != "" && (strings.Contains(containerInfo.Image, "watchducker") || strings.Contains(containerInfo.Image, "naomi233")) {
		return true
	}

	return false
}

// SelfUpdate WatchDucker 自我更新流程
func (su *SelfUpdater) SelfUpdate(ctx context.Context) error {
	logger.Info("开始 WatchDucker 自我更新流程")

	// 步骤1：定位自我容器
	selfContainer, err := su.findSelfContainer(ctx)
	if err != nil {
		return fmt.Errorf("定位自我容器失败: %w", err)
	}

	// 步骤2：获取容器完整配置
	containerConfig, err := su.containerOpsSvc.GetContainerConfig(ctx, selfContainer.ID)
	if err != nil {
		return fmt.Errorf("获取容器配置失败: %w", err)
	}

	// 步骤3：获取新镜像信息
	imageInfo, err := su.containerOpsSvc.GetImageInspect(ctx, selfContainer.Image)
	if err != nil {
		return fmt.Errorf("获取镜像信息失败: %w", err)
	}

	// 步骤4：执行自我更新
	if err := su.selfUpdateContainer(ctx, *selfContainer, containerConfig, imageInfo, selfContainer.Image); err != nil {
		return fmt.Errorf("自我更新失败: %w", err)
	}

	return nil
}

// selfUpdateContainer 执行自我更新的具体步骤
func (su *SelfUpdater) selfUpdateContainer(ctx context.Context, containerInfo types.ContainerInfo, containerConfig *dockerTypes.ContainerJSON, imageInfo *dockerTypes.ImageInspect, newImage string) error {
	// 步骤1：生成随机字符串用于重命名旧容器
	randomSuffix := utils.GenerateRandomString(8)
	oldContainerNewName := fmt.Sprintf("%s_%s", containerInfo.Name, randomSuffix)

	// 步骤2：重命名旧容器（不停止）
	logger.Info("重命名旧容器: %s -> %s", containerInfo.Name, oldContainerNewName)
	if err := su.containerOpsSvc.RenameContainer(ctx, containerInfo.ID, oldContainerNewName); err != nil {
		return fmt.Errorf("重命名旧容器失败: %w", err)
	}

	// 步骤3：使用新镜像启动新容器
	logger.Info("使用新镜像启动新容器: %s", containerInfo.Name)
	newContainerID, err := su.createNewContainer(ctx, containerConfig, imageInfo, newImage, containerInfo.Name)
	if err != nil {
		// 失败时尝试恢复旧容器名称
		su.containerOpsSvc.RenameContainer(ctx, containerInfo.ID, containerInfo.Name)
		return fmt.Errorf("创建新容器失败: %w", err)
	}

	if err := su.containerOpsSvc.StartContainer(ctx, newContainerID); err != nil {
		// 失败时尝试恢复旧容器名称并删除新容器
		su.containerOpsSvc.RenameContainer(ctx, containerInfo.ID, containerInfo.Name)
		su.containerOpsSvc.RemoveContainer(ctx, newContainerID, true)
		return fmt.Errorf("启动新容器失败: %w", err)
	}

	logger.Info("新容器已成功启动，容器ID: %s", newContainerID[:12])

	// 步骤4：停止并删除旧容器
	logger.Info("停止并删除旧容器: %s", oldContainerNewName)
	stopTimeout := 30 * time.Second
	if err := su.containerOpsSvc.StopContainer(ctx, containerInfo.ID, &stopTimeout); err != nil {
		logger.Warn("停止旧容器失败: %v，但新容器已成功启动", err)
		// 继续尝试删除容器
	}

	if err := su.containerOpsSvc.RemoveContainer(ctx, containerInfo.ID, true); err != nil {
		logger.Warn("删除旧容器失败: %v，但新容器已成功启动", err)
	} else {
		logger.Info("旧容器已成功删除")
	}

	logger.Info("WatchDucker 自我更新完成，新容器ID: %s", newContainerID[:12])
	return nil
}

// createNewContainer 使用新镜像创建新容器
func (su *SelfUpdater) createNewContainer(ctx context.Context, containerJSON *dockerTypes.ContainerJSON, imageInfo *dockerTypes.ImageInspect, newImage string, containerName string) (string, error) {
	// 准备创建容器的配置
	config := su.containerSvc.GetCreateConfig(ctx, *containerJSON, imageInfo, newImage)
	hostConfig := su.containerSvc.GetCreateHostConfig(ctx, *containerJSON)
	networkingConfig := su.containerSvc.GetNetworkConfig(ctx, *containerJSON)

	// 仅使用一个网络配置来创建容器，之后再连接其他网络
	simpleNetworkConfig := func() *network.NetworkingConfig {
		oneEndpoint := make(map[string]*network.EndpointSettings)
		for k, v := range networkingConfig.EndpointsConfig {
			oneEndpoint[k] = v
			break
		}
		return &network.NetworkingConfig{EndpointsConfig: oneEndpoint}
	}()

	// 创建新容器
	newContainerID, err := su.containerOpsSvc.CreateContainer(ctx, config, hostConfig, simpleNetworkConfig, containerName)
	if err != nil {
		return "", err
	}

	// 连接其他网络
	if !(hostConfig.NetworkMode.IsHost()) {
		for k := range simpleNetworkConfig.EndpointsConfig {
			err = su.containerOpsSvc.NetworkDisconnect(ctx, k, newContainerID, true)
			if err != nil {
				return "", err
			}
		}

		for k, v := range networkingConfig.EndpointsConfig {
			err = su.containerOpsSvc.NetworkConnect(ctx, k, newContainerID, v)
			if err != nil {
				return "", err
			}
		}
	}

	return newContainerID, nil
}

// Close 关闭所有资源
func (su *SelfUpdater) Close() error {
	if su.clientManager != nil {
		return su.clientManager.Close()
	}
	return nil
}
