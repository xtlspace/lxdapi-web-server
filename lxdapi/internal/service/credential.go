package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"lxdapi/internal/db"
	"lxdapi/models"
)

type ContainerCredentialData struct {
	ContainerName string
	Hash          string
}

func generateContainerHash() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateContainerCredential(containerName string) (*ContainerCredentialData, error) {
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return nil, fmt.Errorf("容器不存在")
	}

	if container.Hash == "" {
		hash, err := generateContainerHash()
		if err != nil {
			return nil, fmt.Errorf("生成容器Hash失败: %v", err)
		}
		if err := db.DB.Model(&container).Update("hash", hash).Error; err != nil {
			return nil, err
		}
		container.Hash = hash
	}

	return &ContainerCredentialData{
		ContainerName: container.Name,
		Hash:          container.Hash,
	}, nil
}

func GetContainerCredential(containerName string) (*ContainerCredentialData, error) {
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return nil, fmt.Errorf("容器凭证不存在")
	}
	return &ContainerCredentialData{
		ContainerName: container.Name,
		Hash:          container.Hash,
	}, nil
}

func GetContainerByHash(hash string) (*ContainerCredentialData, error) {
	var container models.Container
	if err := db.DB.Where("hash = ?", hash).First(&container).Error; err != nil {
		return nil, fmt.Errorf("无效的容器Hash")
	}
	return &ContainerCredentialData{
		ContainerName: container.Name,
		Hash:          container.Hash,
	}, nil
}

func RegenerateContainerHash(containerName string) (*ContainerCredentialData, error) {
	var container models.Container
	if err := db.DB.Where("name = ?", containerName).First(&container).Error; err != nil {
		return nil, fmt.Errorf("容器凭证不存在")
	}

	hash, err := generateContainerHash()
	if err != nil {
		return nil, fmt.Errorf("生成容器Hash失败: %v", err)
	}

	if err := db.DB.Model(&container).Update("hash", hash).Error; err != nil {
		return nil, err
	}
	container.Hash = hash

	return &ContainerCredentialData{
		ContainerName: container.Name,
		Hash:          container.Hash,
	}, nil
}