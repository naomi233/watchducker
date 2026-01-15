package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"watchducker/internal/types"
	"watchducker/pkg/logger"
	"watchducker/pkg/utils"

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
func (cs *ContainerService) StopContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	cli := cs.clientManager.GetClient()

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
func (cs *ContainerService) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	cli := cs.clientManager.GetClient()

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

// RenameContainer 重命名容器
func (cs *ContainerService) RenameContainer(ctx context.Context, containerID, newName string) error {
	cli := cs.clientManager.GetClient()

	logger.Debug("正在重命名容器: %s -> %s", containerID[:12], newName)

	if err := cli.ContainerRename(ctx, containerID, newName); err != nil {
		logger.Error("重命名容器 %s 失败: %v", containerID[:12], err)
		return fmt.Errorf("重命名容器 %s 失败: %w", containerID[:12], err)
	}

	logger.Debug("容器 %s 已成功重命名为 %s", containerID[:12], newName)
	return nil
}

// StartContainer 启动容器
func (cs *ContainerService) StartContainer(ctx context.Context, containerID string) error {
	cli := cs.clientManager.GetClient()

	logger.Debug("正在启动容器: %s", containerID[:12])

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		logger.Error("启动容器 %s 失败: %v", containerID[:12], err)
		return fmt.Errorf("启动容器 %s 失败: %w", containerID[:12], err)
	}

	logger.Debug("容器 %s 已成功启动", containerID[:12])
	return nil
}

// CreateContainer 创建容器
func (cs *ContainerService) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error) {
	cli := cs.clientManager.GetClient()

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
func (cs *ContainerService) GetContainerConfig(ctx context.Context, containerID string) (*dockerTypes.ContainerJSON, error) {
	cli := cs.clientManager.GetClient()

	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		logger.Error("获取容器 %s 配置失败: %v", containerID[:12], err)
		return nil, fmt.Errorf("获取容器 %s 配置失败: %w", containerID[:12], err)
	}

	return &containerJSON, nil
}

func (cs *ContainerService) GetImageInspect(ctx context.Context, imageName string) (*dockerTypes.ImageInspect, error) {
	cli := cs.clientManager.GetClient()

	imageInfo, _, err := cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		logger.Error("获取镜像 %s 信息失败: %v", imageName, err)
		return nil, fmt.Errorf("获取镜像 %s 信息失败: %w", imageName, err)
	}

	return &imageInfo, nil
}

func (cs *ContainerService) GetCreateConfig(ctx context.Context, containerJSON dockerTypes.ContainerJSON, imageInfo *dockerTypes.ImageInspect, imageName string) *container.Config {
	config := containerJSON.Config
	hostConfig := containerJSON.HostConfig
	imageConfig := imageInfo.Config

	if config.WorkingDir == imageConfig.WorkingDir {
		config.WorkingDir = ""
	}

	if config.User == imageConfig.User {
		config.User = ""
	}

	if hostConfig.NetworkMode.IsContainer() {
		config.Hostname = ""
	}

	if utils.SliceEqual(config.Entrypoint, imageConfig.Entrypoint) {
		config.Entrypoint = nil
		if utils.SliceEqual(config.Cmd, imageConfig.Cmd) {
			config.Cmd = nil
		}
	}

	// Clear HEALTHCHECK configuration (if default)
	if config.Healthcheck != nil && imageConfig.Healthcheck != nil {
		if utils.SliceEqual(config.Healthcheck.Test, imageConfig.Healthcheck.Test) {
			config.Healthcheck.Test = nil
		}

		if config.Healthcheck.Retries == imageConfig.Healthcheck.Retries {
			config.Healthcheck.Retries = 0
		}

		if config.Healthcheck.Interval == imageConfig.Healthcheck.Interval {
			config.Healthcheck.Interval = 0
		}

		if config.Healthcheck.Timeout == imageConfig.Healthcheck.Timeout {
			config.Healthcheck.Timeout = 0
		}

		if config.Healthcheck.StartPeriod == imageConfig.Healthcheck.StartPeriod {
			config.Healthcheck.StartPeriod = 0
		}
	}

	config.Env = utils.SliceSubtract(config.Env, imageConfig.Env)

	config.Labels = utils.StringMapSubtract(config.Labels, imageConfig.Labels)

	config.Volumes = utils.StructMapSubtract(config.Volumes, imageConfig.Volumes)

	// 从容器中去除镜像中暴露的端口
	for k := range config.ExposedPorts {
		if _, ok := imageConfig.ExposedPorts[k]; ok {
			delete(config.ExposedPorts, k)
		}
	}
	for p := range containerJSON.HostConfig.PortBindings {
		config.ExposedPorts[p] = struct{}{}
	}

	config.Image = imageName
	return config
}

func (cs *ContainerService) GetCreateHostConfig(ctx context.Context, containerJSON dockerTypes.ContainerJSON) *container.HostConfig {
	hostConfig := containerJSON.HostConfig

	for i, link := range hostConfig.Links {
		name := link[0:strings.Index(link, ":")]
		alias := link[strings.LastIndex(link, "/"):]

		hostConfig.Links[i] = fmt.Sprintf("%s:%s", name, alias)
	}

	return hostConfig
}

func (cs *ContainerService) GetNetworkConfig(ctx context.Context, containerJSON dockerTypes.ContainerJSON) *network.NetworkingConfig {
	config := &network.NetworkingConfig{
		EndpointsConfig: containerJSON.NetworkSettings.Networks,
	}

	// Remove the old container ID alias from the network aliases, as it would accumulate across updates otherwise
	for _, ep := range config.EndpointsConfig {
		cidAlias := containerJSON.ID[:12]
		aliases := make([]string, 0, len(ep.Aliases))

		for _, alias := range ep.Aliases {
			if alias == cidAlias {
				continue
			}
			aliases = append(aliases, alias)
		}

		ep.Aliases = aliases
	}

	return config
}

func (cs *ContainerService) NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error {
	cli := cs.clientManager.GetClient()

	if err := cli.NetworkDisconnect(ctx, networkID, containerID, force); err != nil {
		logger.Error("断开容器 %s 与网络 %s 的连接失败: %v", containerID[:12], networkID, err)
		return fmt.Errorf("断开容器 %s 与网络 %s 的连接失败: %w", containerID[:12], networkID, err)
	}

	return nil
}

func (cs *ContainerService) NetworkConnect(ctx context.Context, networkID, containerID string, endpointConfig *network.EndpointSettings) error {
	cli := cs.clientManager.GetClient()

	if err := cli.NetworkConnect(ctx, networkID, containerID, endpointConfig); err != nil {
		logger.Error("连接容器 %s 到网络 %s 失败: %v", containerID[:12], networkID, err)
		return fmt.Errorf("连接容器 %s 到网络 %s 失败: %w", containerID[:12], networkID, err)
	}

	return nil
}
