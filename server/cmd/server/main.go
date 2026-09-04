// 电气车间隐患闭环系统后端入口。
package main

import (
	"log"

	"github.com/joho/godotenv"

	"hazard-system/server/internal/auth"
	"hazard-system/server/internal/config"
	"hazard-system/server/internal/database"
	"hazard-system/server/internal/handler"
	"hazard-system/server/internal/repo"
	"hazard-system/server/internal/router"
	"hazard-system/server/internal/service"
	"hazard-system/server/internal/upload"
)

func main() {
	// .env 可选：存在则加载，不存在时使用进程环境变量。
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置错误: %v", err)
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if err := database.SeedIfEmpty(db); err != nil {
		log.Fatalf("示例数据写入失败: %v", err)
	}

	am := auth.NewManager(cfg.JWTSecret, cfg.JWTTokenTTL, cfg.AdminPassword)

	store, err := upload.NewStore(cfg.UploadDir, db)
	if err != nil {
		log.Fatalf("图片存储初始化失败: %v", err)
	}

	// 组装数据访问与业务层。
	hazardRepo := repo.NewHazardRepo(db)
	unitRepo := repo.NewUnitRepo(db)
	typeRepo := repo.NewHazardTypeRepo(db)
	imageRepo := repo.NewImageRepo(db)

	hazardSvc := service.NewHazardService(hazardRepo, unitRepo, typeRepo, imageRepo)
	unitSvc := service.NewUnitService(unitRepo, hazardRepo)
	typeSvc := service.NewTypeService(typeRepo, hazardRepo)
	imageSvc := service.NewImageService(imageRepo, store)

	srv := handler.NewServer(hazardSvc, unitSvc, typeSvc, imageSvc, am, store, cfg.MaxUploadBytes)

	engine := router.NewRouter(srv, am)

	log.Printf("隐患闭环系统服务启动，监听 :%s", cfg.ServerPort)
	if err := engine.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
