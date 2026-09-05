package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"net/http"
	"lxdapi/handlers"
	"lxdapi/internal/api/admin"
	"lxdapi/internal/api/console"
	"lxdapi/internal/api/container"
	"lxdapi/internal/api/public"
	"lxdapi/internal/api/system"
	"lxdapi/internal/core"
	"lxdapi/internal/db"
	"lxdapi/internal/executor"
	"lxdapi/internal/ipv4"
	"lxdapi/internal/ipv6"
	"lxdapi/internal/middleware"
	"lxdapi/internal/service"
	"lxdapi/internal/task"
	"lxdapi/internal/monitor"
	"lxdapi/models"
	"lxdapi/pkg/logger"
	"lxdapi/pkg/plugin"
	"lxdapi/pkg/response"
	sysinfo "lxdapi/pkg/system"
	tlsManager "lxdapi/pkg/tls"
	"lxdapi/pkg/version"
	"lxdapi/plugins/nginx"
	"log"
	"os"
	"path/filepath"
)

//go:embed templates static
var embeddedFiles embed.FS

func main() {
	if err := core.LoadConfig("configs/config.yaml"); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	cfg := core.GlobalConfig

	logger.Init(cfg.System.Logger.Level, cfg.System.Logger.Colorful)
	logger.Info("配置加载成功")

	if err := validateSecurityConfig(cfg); err != nil {
		logger.Error("安全配置校验失败: %v", err)
		return
	}

	if err := db.Init(); err != nil {
		logger.Error("数据库初始化失败: %v", err)
		return
	}
	logger.OK("数据库连接成功")

	logger.Info("正在初始化网络组件...")
	
	if err := ipv4.FlushNftTables(); err != nil {
		logger.Error("清空nftables表失败: %v", err)
	} else {
		logger.OK("nftables 表已清空，等待按数据库重建")
	}
	
	if err := ipv4.InitManager(); err != nil {
		logger.Error("IPv4管理器初始化失败: %v", err)
	}

	if err := ipv6.InitManager(); err != nil {
		logger.Error("IPv6管理器初始化失败: %v", err)
	}
	
	var portMappingV4Count int64
	db.DB.Model(&models.PortMappingV4{}).Count(&portMappingV4Count)
	logger.OK("IPv4端口映射: %d条规则", portMappingV4Count)
	
	var portMappingV6Count int64
	db.DB.Model(&models.PortMappingV6{}).Count(&portMappingV6Count)
	logger.OK("IPv6端口映射: %d条规则", portMappingV6Count)
	


	if err := monitor.InitMonitor(); err != nil {
		logger.Error("流量监控器初始化失败: %v", err)
	} else {
		logger.OK("流量监控器初始化成功")
	}

	pluginManager := plugin.InitManager()
	logger.OK("插件管理器初始化成功")

	if cfg.Plugins.Nginx.Enabled {
		nginxPlugin := nginx.NewNginxPlugin()
		if err := pluginManager.Register(nginxPlugin); err != nil {
			logger.Error("注册 Nginx 插件失败: %v", err)
		} else {
			logger.OK("Nginx 插件注册成功")
		}
	} else {
		logger.Info("Nginx 插件已禁用，跳过注册")
	}

		
	if err := pluginManager.StartAll(); err != nil {
		logger.Error("插件启动失败: %v", err)
	} else {
		logger.OK("所有插件启动成功")
	}

	containerService := service.NewContainerService()
	ipv4Service := service.NewIPv4Service()

	system.InitContainerService(containerService)
	system.InitIPv4Service(ipv4Service)
	container.InitContainerService(containerService)
	admin.InitBrandService()

	executor.ClearPendingTasks()

	if err := executor.InitQueue(); err != nil {
		logger.Error("任务队列初始化失败: %v", err)
	} else {
		logger.OK("任务队列初始化成功")
	}

	ctx := context.Background()
	if monitor.GlobalMonitor != nil {
		go monitor.GlobalMonitor.Start(ctx)
		logger.OK("流量监控器已启动")
	}

	task.StartAutoCleanup()
	logger.OK("自动清理任务已启动")

	go func() {
		logger.Info("启动时自动同步镜像模板...")
		svc := service.NewTemplateService()
		added, updated, deleted, err := svc.SyncFromLXD(context.Background())
		if err != nil {
			logger.Error("启动时同步模板失败: %v", err)
			return
		}
		logger.OK("启动模板同步完成: 新增 %d, 更新 %d, 删除 %d", added, updated, deleted)
	}()

	gin.SetMode(cfg.System.Server.Mode)
	r := gin.Default()

	store := cookie.NewStore([]byte(cfg.Admin.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: 2,
	})
	r.Use(sessions.Sessions("lxdapi_session", store))
	r.Use(middleware.BrandMiddleware())

	tmpl := template.New("").Funcs(template.FuncMap{})
	
	templatesFS, err := fs.Sub(embeddedFiles, "templates")
	if err != nil {
		logger.Error("加载嵌入式模板失败: %v", err)
		patterns := []string{
			"templates/*.html",
			"templates/admin/*.html",
			"templates/container/*.html",
		}
		
		for _, pattern := range patterns {
			files, err := filepath.Glob(pattern)
			if err != nil {
				logger.Error("加载模板失败 %s: %v", pattern, err)
				continue
			}
			for _, file := range files {
				name := filepath.ToSlash(file[len("templates/"):])
				content, err := os.ReadFile(file)
				if err != nil {
					logger.Error("读取模板失败 %s: %v", file, err)
					continue
				}
				_, err = tmpl.New(name).Parse(string(content))
				if err != nil {
					logger.Error("解析模板失败 %s: %v", file, err)
				}
			}
		}
	} else {
		fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".html" {
				return nil
			}
			
			content, err := fs.ReadFile(templatesFS, path)
			if err != nil {
				logger.Error("读取嵌入式模板失败 %s: %v", path, err)
				return nil
			}
			
			name := filepath.ToSlash(path)
			_, err = tmpl.New(name).Parse(string(content))
			if err != nil {
				logger.Error("解析嵌入式模板失败 %s: %v", path, err)
			}
			return nil
		})
	}
	
	r.SetHTMLTemplate(tmpl)
	logger.OK("模板加载完成")
	
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(302, "/")
	})

	pluginManager.RegisterRoutes(r)

	r.GET("/", func(c *gin.Context) {
		sysInfo := sysinfo.GetSystemInfo()
		response.Success(c, gin.H{
			"author":       "xkatld",
			"project":      "https://github.com/xkatld/lxdapi-web-server",
			"description":  "主流财务系统对接支持，提供完整的Web管理界面与RESTful API",
			"version":      version.Version,
			"name":         sysInfo.Name,
			"os":           sysInfo.OS,
			"arch":         sysInfo.Arch,
			"lxd_version":  sysInfo.LXDVersion,
			"distribution": sysInfo.Distribution,
			"kernel":       sysInfo.Kernel,
		})
	})
	
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err == nil {
		r.StaticFS("/static", http.FS(staticFS))
	}

	r.GET("/console", handlers.ConsolePage)
	r.GET("/ws/console", console.HandleWebSocket)
	r.GET("/container/dashboard/lite", handlers.ContainerDashboardLite)
	r.GET("/container/dashboard/base", handlers.ContainerDashboardBase)
	
	r.GET("/admin/login", handlers.AdminLogin)
	r.GET("/api/admin/captcha", admin.GetCaptcha)
	r.POST("/api/admin/login", admin.Login)
	
	adminPages := r.Group("/admin")
	adminPages.Use(middleware.AdminPageAuth())
	{
		adminPages.GET("/dashboard", handlers.AdminDashboard)
		adminPages.GET("/containers", handlers.AdminContainers)
		adminPages.GET("/containers/:name", handlers.AdminContainerDetail)
		adminPages.GET("/tasks", handlers.AdminTasks)
		adminPages.GET("/ip-pool/v4", handlers.AdminIPPoolV4)
		adminPages.GET("/port-mapping/v4", handlers.AdminPortMappingV4)
		adminPages.GET("/port-mapping/v6", handlers.AdminPortMappingV6)
		adminPages.GET("/templates", handlers.AdminTemplates)
		adminPages.GET("/brand-settings", handlers.AdminBrandSettings)
		adminPages.GET("/nginx", handlers.AdminNginx)
		adminPages.GET("/nserIPv6", handlers.AdminIPv6Neighbor)
	}
	
	r.POST("/api/admin/logout", admin.Logout)
	r.GET("/admin/logout", handlers.AdminLogout)
	
	r.GET("/container/login", handlers.ContainerLogin)
	r.GET("/container/dashboard", handlers.ContainerDashboard)

	systemAPI := r.Group("/api/system")
	systemAPI.Use(middleware.SystemAuth())
	{
		systemAPI.POST("/containers", system.CreateContainer)
		systemAPI.GET("/containers", system.ListContainers)
		systemAPI.GET("/containers/:name", system.GetContainer)
		systemAPI.DELETE("/containers/:name", system.DeleteContainer)
		systemAPI.POST("/containers/:name/action", system.ContainerAction)
		systemAPI.PUT("/containers/:name/config", system.UpdateContainerConfig)
		systemAPI.POST("/port-mapping/allocate", system.AllocatePortMapping)
		systemAPI.POST("/port-mapping/release", system.ReleasePortMapping)
		systemAPI.GET("/port-mapping", system.ListPortMappings)
		systemAPI.GET("/ip", system.GetContainerIP)
		systemAPI.POST("/ip/allocate", system.AllocateIP)
		systemAPI.POST("/ip/release", system.ReleaseIP)
		systemAPI.GET("/traffic", system.GetTraffic)
		systemAPI.POST("/traffic/reset", system.ResetTraffic)
		systemAPI.GET("/tasks", system.ListTasks)
		systemAPI.GET("/tasks/detail", system.GetTask)
		systemAPI.GET("/containers/:name/credential", system.GetContainerCredential)
		systemAPI.POST("/containers/:name/credential/regenerate", system.RegenerateContainerCredential)
		systemAPI.POST("/console/create-token", console.CreateToken)
		systemAPI.GET("/admin/access-token", system.GetAdminAccessToken)
	}

	r.GET("/api/public/port-range-config", admin.GetPortRangeConfigPublic)

	adminAPI := r.Group("/api/admin")
	adminAPI.Use(middleware.AdminAuth())
	{
		adminAPI.GET("/dashboard", admin.GetDashboard)
		adminAPI.GET("/host/stats", admin.GetHostStats)
		adminAPI.GET("/containers", admin.ListContainers)
		adminAPI.GET("/containers/:name", admin.GetContainer)
		adminAPI.DELETE("/containers/:name", admin.DeleteContainer)
		adminAPI.POST("/containers/:name/action", admin.ContainerAction)
		adminAPI.GET("/containers/:name/config", admin.GetContainerConfig)
		adminAPI.PUT("/containers/:name/config", admin.UpdateContainerConfig)
		adminAPI.PUT("/containers/:name/remark", admin.UpdateContainerRemark)
		adminAPI.POST("/containers/create", system.CreateContainer)
		adminAPI.GET("/containers/:name/credential", admin.GetContainerCredential)
		adminAPI.POST("/containers/:name/credential", admin.RegenerateContainerCredential)
		adminAPI.GET("/containers/:name/ip", admin.GetContainerIP)
		adminAPI.POST("/containers/:name/ip/allocate", admin.AllocateContainerIP)
		adminAPI.GET("/cache/containers", admin.GetContainersCache)
		adminAPI.GET("/tasks", admin.GetTaskList)
		adminAPI.GET("/tasks/:id", admin.GetTask)
		adminAPI.DELETE("/tasks/:id", admin.DeleteTask)
		adminAPI.POST("/tasks/batch-delete", admin.BatchDeleteTasks)
		adminAPI.GET("/port-mapping", admin.ListPortMappings)
		adminAPI.POST("/port-mapping/allocate", admin.AllocatePortMapping)
		adminAPI.POST("/port-mapping/release", admin.ReleasePortMapping)
		adminAPI.GET("/nat-config", admin.GetNATConfig)
		adminAPI.POST("/nat-config", admin.SaveNATConfig)
		adminAPI.GET("/ip", admin.GetIPList)
		adminAPI.POST("/ip/allocate", admin.AllocateIP)
		adminAPI.POST("/ip/release", admin.ReleaseIP)
		adminAPI.GET("/ip/pool", admin.GetIPPool)
		adminAPI.POST("/ip/pool", admin.AddIPToPool)
		adminAPI.PUT("/ip/pool", admin.UpdateIPPool)
		adminAPI.DELETE("/ip/pool", admin.DeleteIPFromPool)
		adminAPI.POST("/ip/pool/batch", admin.BatchAddIPToPool)
		adminAPI.POST("/ip/pool/batch-delete", admin.BatchDeleteIPFromPool)
		adminAPI.GET("/templates", admin.GetTemplateList)
		adminAPI.POST("/templates/sync", admin.SyncTemplates)
		adminAPI.DELETE("/templates/:fingerprint", admin.DeleteTemplate)
		adminAPI.POST("/templates/batch-delete", admin.BatchDeleteTemplates)
		adminAPI.GET("/brand-settings", admin.GetBrandSettings)
		adminAPI.POST("/brand-settings", admin.UpdateBrandSettings)
		adminAPI.GET("/port-range/config", admin.GetPortRangeConfig)
		adminAPI.POST("/port-range/config", admin.SavePortRangeConfig)
		adminAPI.GET("/nserIPv6/settings", admin.GetIPv6NeighborConfig)
		adminAPI.PUT("/nserIPv6/settings", admin.SaveIPv6NeighborConfig)
		adminAPI.GET("/ip-pool/settings", admin.GetIPPoolSettings)
		adminAPI.PUT("/ip-pool/settings", admin.UpdateIPPoolSettings)
		adminAPI.POST("/console/create-token", console.CreateToken)
		adminAPI.GET("/network/nat", admin.GetNetworkNATStatus)
		adminAPI.POST("/network/nat", admin.SetNetworkNATStatus)
	}

	r.GET("/api/public/brand-settings", public.GetBrandSettings)
	r.GET("/api/public/ip-pool-settings", admin.GetIPPoolSettingsPublic)
	
	r.GET("/api/container/captcha", container.GetCaptcha)
	r.POST("/api/container/verify", container.VerifyAccess)
	
	containerAPI := r.Group("/api/container")
	containerAPI.Use(middleware.ContainerAuth())
	{
		containerAPI.GET("/info", container.GetInfo)
		containerAPI.GET("/templates", container.GetTemplateList)
		containerAPI.POST("/action", container.Action)
		containerAPI.POST("/port-mapping/allocate", container.AllocatePortMapping)
		containerAPI.POST("/port-mapping/release", container.ReleasePortMapping)
		containerAPI.GET("/port-mapping", container.ListPortMappings)
		containerAPI.GET("/ip", container.GetIP)
		containerAPI.POST("/ip/allocate", container.AllocateIP)
		containerAPI.POST("/ip/release", container.ReleaseIP)
		containerAPI.POST("/regenerate-hash", container.RegenerateHash)
		containerAPI.POST("/console/create-token", console.CreateToken)
	}

	addr := fmt.Sprintf("%s:%d", cfg.System.Server.Host, cfg.System.Server.Port)
	
	if cfg.System.Server.TLS.Enabled {
		startWithTLS(r, addr, cfg)
	} else {
		logger.Info("服务启动 (HTTP)，监听 %s", addr)
		if err := r.Run(addr); err != nil {
			logger.Error("服务启动失败: %v", err)
		}
	}
}

// validateSecurityConfig 校验关键安全配置，避免使用占位符或空值导致认证失效
func validateSecurityConfig(cfg *core.Config) error {
	if cfg.System.Security.APIHash == "" || cfg.System.Security.APIHash == "__API_HASH__" {
		return fmt.Errorf("安全配置 api_hash 未正确设置（不能为空或占位符）")
	}
	if cfg.Admin.SessionSecret == "" || cfg.Admin.SessionSecret == "__SESSION_SECRET__" {
		return fmt.Errorf("安全配置 session_secret 未正确设置（不能为空或占位符）")
	}
	return nil
}

func startWithTLS(r *gin.Engine, addr string, cfg *core.Config) {
	tlsCfg := cfg.System.Server.TLS
	certMgr := tlsManager.NewCertificateManager(tlsCfg.CertFile, tlsCfg.KeyFile)
	
	certFile := tlsCfg.CertFile
	keyFile := tlsCfg.KeyFile
	certSource := ""

	if tlsCfg.AutoGenerate {
		generateSelfSignedCert(certMgr)
		certSource = "自签名"
	}
	
	if certSource == "" {
		log.Fatalf("无可用证书，无法启动HTTPS服务")
	}
	
	logger.Info("证书来源: %s", certSource)
	logger.Info("服务启动 (HTTPS)，监听 %s", addr)
	logger.Info("TLS证书: %s", certFile)
	logger.Info("TLS私钥: %s", keyFile)
	
	if err := r.RunTLS(addr, certFile, keyFile); err != nil {
		logger.Error("HTTPS服务器启动失败: %v", err)
	}
}

func generateSelfSignedCert(certMgr *tlsManager.CertificateManager) {
	if !certMgr.CertificateExists() {
		logger.Info("证书文件不存在，开始生成自签名证书...")
		opts := tlsManager.GenerateOptions{
			Organization:  "lxdapi",
			Country:       "US",
			Province:      "State",
			Locality:      "City",
			ServerIPs:     []string{"127.0.0.1"},
			ServerDomains: []string{"localhost"},
			ValidityDays:  3650,
		}
		
		if err := certMgr.GenerateSelfSignedCert(opts); err != nil {
			logger.Error("生成自签名证书失败: %v", err)
			log.Fatalf("生成自签名证书失败: %v", err)
		}
		
		logger.OK("自签名证书生成成功(有效期3650天)")
	} else if !certMgr.ValidateCertificate() {
		logger.Warn("现有证书无效或已过期，重新生成...")
		opts := tlsManager.GenerateOptions{
			Organization:  "lxdapi",
			Country:       "US",
			Province:      "State",
			Locality:      "City",
			ServerIPs:     []string{"127.0.0.1"},
			ServerDomains: []string{"localhost"},
			ValidityDays:  3650,
		}
		
		if err := certMgr.GenerateSelfSignedCert(opts); err != nil {
			logger.Error("重新生成证书失败: %v", err)
			log.Fatalf("重新生成证书失败: %v", err)
		}
		
		logger.OK("证书重新生成成功 (有效期3650天)")
	} else {
		logger.OK("使用现有证书")
	}
}
