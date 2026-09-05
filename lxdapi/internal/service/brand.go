package service

import (
	"lxdapi/internal/db"
	"lxdapi/models"
	"lxdapi/pkg/logger"
)

type BrandService struct{}

func NewBrandService() *BrandService {
	return &BrandService{}
}

func (s *BrandService) GetSettings() (*models.BrandSettings, error) {
	var settings models.BrandSettings

	if err := db.DB.First(&settings).Error; err != nil {
		settings = models.BrandSettings{
			AdminSystemName:         "LXD API - 管理后台",
			AdminSystemTitle:        "管理后台 - LXD容器管理系统",
			AdminLoginTitle:         "管理员登录",
			AdminBgImage:            "",
			AdminBgOpacity:          75,
			AdminContentOpacity:     85,
			ContainerSystemName:     "LXD API - 容器控制",
			ContainerSystemTitle:    "容器管理 - LXD容器管理系统",
			ContainerLoginTitle:     "容器登录",
			ContainerBgImage:        "",
			ContainerBgOpacity:      75,
			ContainerContentOpacity: 85,
			ContainerNotice:         "",
		}

		if err := db.DB.Create(&settings).Error; err != nil {
			logger.Error("创建默认品牌设置失败: %v", err)
			return nil, err
		}
		logger.Info("已创建默认品牌设置")
	}

	return &settings, nil
}

func (s *BrandService) UpdateSettings(settings *models.BrandSettings) error {
	var existing models.BrandSettings

	if err := db.DB.First(&existing).Error; err != nil {
		if err := db.DB.Create(settings).Error; err != nil {
			logger.Error("创建品牌设置失败: %v", err)
			return err
		}
		logger.OK("品牌设置创建成功")
		return nil
	}

	existing.AdminSystemName = settings.AdminSystemName
	existing.AdminSystemTitle = settings.AdminSystemTitle
	existing.AdminLoginTitle = settings.AdminLoginTitle
	existing.AdminBgImage = settings.AdminBgImage
	existing.AdminBgOpacity = settings.AdminBgOpacity
	existing.AdminContentOpacity = settings.AdminContentOpacity
	existing.ContainerSystemName = settings.ContainerSystemName
	existing.ContainerSystemTitle = settings.ContainerSystemTitle
	existing.ContainerLoginTitle = settings.ContainerLoginTitle
	existing.ContainerBgImage = settings.ContainerBgImage
	existing.ContainerBgOpacity = settings.ContainerBgOpacity
	existing.ContainerContentOpacity = settings.ContainerContentOpacity
	existing.ContainerNotice = settings.ContainerNotice
	existing.ContainerNoticeOpacity = settings.ContainerNoticeOpacity

	if err := db.DB.Save(&existing).Error; err != nil {
		logger.Error("更新品牌设置失败: %v", err)
		return err
	}

	logger.OK("品牌设置更新成功")
	return nil
}

func (s *BrandService) ResetToDefault() error {
	var settings models.BrandSettings

	if err := db.DB.First(&settings).Error; err != nil {
		return err
	}

	settings.AdminSystemName = "LXD API - 管理后台"
	settings.AdminSystemTitle = "管理后台 - LXD容器管理系统"
	settings.AdminLoginTitle = "管理员登录"
	settings.AdminBgImage = ""
	settings.AdminBgOpacity = 75
	settings.AdminContentOpacity = 85
	settings.ContainerSystemName = "LXD API - 容器控制"
	settings.ContainerSystemTitle = "容器管理 - LXD容器管理系统"
	settings.ContainerLoginTitle = "容器登录"
	settings.ContainerBgImage = ""
	settings.ContainerBgOpacity = 75
	settings.ContainerContentOpacity = 85
	settings.ContainerNotice = ""

	if err := db.DB.Save(&settings).Error; err != nil {
		logger.Error("重置品牌设置失败: %v", err)
		return err
	}

	logger.OK("品牌设置已重置为默认值")
	return nil
}
