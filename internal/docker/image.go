package docker

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"watchducker/internal/types"
	"watchducker/pkg/logger"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

// ImageService 镜像服务
type ImageService struct {
	clientManager *ClientManager
}

// NewImageService 创建镜像服务实例
func NewImageService(clientManager *ClientManager) *ImageService {
	return &ImageService{
		clientManager: clientManager,
	}
}

// getImageList 获取镜像列表的通用方法
func (is *ImageService) getImageList(ctx context.Context, imageName string) ([]image.Summary, error) {
	cli := is.clientManager.GetClient()

	filter := filters.NewArgs()
	filter.Add("reference", imageName)

	return cli.ImageList(ctx, image.ListOptions{
		Filters: filter,
	})
}

// GetLocalHash 获取本地镜像的哈希值
func (is *ImageService) GetLocalHash(ctx context.Context, imageName string) (string, error) {
	images, err := is.getImageList(ctx, imageName)
	if err != nil {
		return "", fmt.Errorf("获取本地镜像列表失败: %w", err)
	}

	if len(images) == 0 {
		return "", fmt.Errorf("本地不存在镜像: %s", imageName)
	}

	// 使用镜像ID作为哈希值
	return images[0].ID, nil
}

// GetRemoteHash 获取远程镜像的哈希值
func (is *ImageService) GetRemoteHash(ctx context.Context, imageName string) (string, error) {
	cli := is.clientManager.GetClient()

	// 拉取镜像以获取最新信息
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("拉取镜像失败: %w", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		// 输出拉取镜像日志
		logger.Debug("%s", scanner.Text())
	}

	// 重新获取镜像信息以获取最新的哈希值
	images, err := is.getImageList(ctx, imageName)
	if err != nil {
		return "", fmt.Errorf("获取更新后的镜像信息失败: %w", err)
	}

	if len(images) == 0 {
		return "", fmt.Errorf("拉取后未找到镜像: %s", imageName)
	}

	return images[0].ID, nil
}

// CheckUpdate 检查镜像是否有更新
func (is *ImageService) CheckUpdate(ctx context.Context, imageName string) (*types.ImageCheckResult, error) {
	result := &types.ImageCheckResult{
		Name:      imageName,
		CheckedAt: time.Now(),
	}

	// 获取本地镜像哈希
	localHash, err := is.GetLocalHash(ctx, imageName)
	if err != nil {
		result.Error = fmt.Sprintf("获取本地镜像信息失败: %v", err)
		return result, err
	}
	result.LocalHash = localHash

	// 获取远程镜像哈希
	remoteHash, err := is.GetRemoteHash(ctx, imageName)
	if err != nil {
		result.Error = fmt.Sprintf("获取远程镜像信息失败: %v", err)
		return result, err
	}
	result.RemoteHash = remoteHash

	// 比较哈希值判断是否有更新
	result.IsUpdated = localHash != remoteHash

	return result, nil
}

// CleanDanglingImages 清理悬空镜像
func (is *ImageService) CleanDanglingImages(ctx context.Context) error {
	cli := is.clientManager.GetClient()

	report, err := cli.ImagesPrune(ctx, filters.NewArgs(
		filters.Arg("dangling", "true"),
	))
	logger.Debug("悬空镜像清理报告: %+v", report)
	if err != nil {
		return fmt.Errorf("清理悬空镜像失败: %w", err)
	}

	return nil
}
