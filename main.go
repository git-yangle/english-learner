package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"english-learner/internal/handler"
	jsonrepo "english-learner/internal/repository/json"
	"english-learner/internal/service"
)

// webFS 内嵌前端静态文件到 binary
//
//go:embed web
var webFS embed.FS

func main() {
	// 1. 初始化数据目录（不存在则创建）
	if err := initDataDirs(); err != nil {
		log.Fatalf("初始化数据目录失败: %v", err)
	}

	// 2. 初始化 Repository 层（JSON 文件实现）
	libraryRepo := jsonrepo.NewLibraryRepo("data/libraries")
	planRepo := jsonrepo.NewPlanRepo(filepath.Join("data", "user", "plan.json"))
	recordRepo := jsonrepo.NewRecordRepo(filepath.Join("data", "user", "records.json"))
	checkInRepo := jsonrepo.NewCheckInRepo(filepath.Join("data", "user", "checkins.json"))

	// 加载词库到内存
	if err := libraryRepo.Reload(); err != nil {
		log.Fatalf("加载词库失败: %v", err)
	}

	// 3. 初始化 Service 层（依赖注入）
	librarySvc := service.NewLibraryService(libraryRepo)
	reviewSvc := service.NewReviewService(recordRepo)
	planSvc := service.NewPlanService(planRepo, libraryRepo, reviewSvc)
	studySvc := service.NewStudyService(planSvc, libraryRepo, recordRepo, reviewSvc)
	checkInSvc := service.NewCheckInService(checkInRepo, recordRepo, planRepo, reviewSvc)
	statsSvc := service.NewStatsService(checkInRepo, recordRepo, libraryRepo)

	// 4. 初始化 Handler 层并注册路由
	router := handler.NewRouter(librarySvc, planSvc, studySvc, checkInSvc, statsSvc, webFS)
	engine := gin.Default()
	router.Register(engine)

	// 5. 打印启动信息
	fmt.Println("English Learner 启动成功，访问: http://localhost:8080")

	// 6. 启动 HTTP 服务，监听 :8080
	log.Fatal(engine.Run(":8080"))
}

// initDataDirs 初始化数据目录（不存在则创建）
func initDataDirs() error {
	dirs := []string{
		"data/libraries",
		"data/user",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}
	return nil
}
