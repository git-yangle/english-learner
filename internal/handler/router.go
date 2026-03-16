package handler

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"

	"english-learner/internal/service"
)

// Router 封装所有 HTTP handler，持有各业务服务引用
type Router struct {
	librarySvc service.LibraryService
	planSvc    service.PlanService
	studySvc   service.StudyService
	checkInSvc service.CheckInService
	statsSvc   service.StatsService
	// webFS 前端静态文件 embed FS（由 main.go 注入）
	webFS embed.FS
}

// NewRouter 创建路由实例，注入所有业务服务和前端静态文件
func NewRouter(
	librarySvc service.LibraryService,
	planSvc service.PlanService,
	studySvc service.StudyService,
	checkInSvc service.CheckInService,
	statsSvc service.StatsService,
	webFS embed.FS,
) *Router {
	return &Router{
		librarySvc: librarySvc,
		planSvc:    planSvc,
		studySvc:   studySvc,
		checkInSvc: checkInSvc,
		statsSvc:   statsSvc,
		webFS:      webFS,
	}
}

// Register 注册所有路由到 gin engine
func (r *Router) Register(engine *gin.Engine) {
	// 注册前端静态文件（embed 方式内嵌到 binary）
	webSubFS, _ := fs.Sub(r.webFS, "web")
	engine.StaticFS("/web", http.FS(webSubFS))
	// 根路径重定向到前端入口
	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/index.html")
	})

	// API 路由分组
	api := engine.Group("/api")

	// 词库相关路由
	api.GET("/libraries", r.listLibraries)
	api.GET("/libraries/:id", r.getLibrary)
	api.POST("/libraries/reload", r.reloadLibraries)

	// 学习计划相关路由
	api.POST("/plan/init", r.initPlan)
	api.GET("/plan/today", r.getTodayPlan)
	api.GET("/plan/overview", r.getPlanOverview)

	// 学习相关路由
	api.GET("/study/browse", r.browse)
	api.POST("/study/browse/mark", r.markBrowse)
	api.GET("/study/quiz", r.generateQuiz)
	api.POST("/study/quiz/answer", r.submitQuizAnswer)
	api.GET("/study/dictation", r.getDictationWord)
	api.POST("/study/dictation/answer", r.submitDictation)

	// 打卡相关路由
	api.GET("/checkin/score", r.calcScore)
	api.POST("/checkin", r.doCheckIn)
	api.GET("/checkin/stats", r.getCheckInStats)

	// 统计相关路由
	api.GET("/stats/overview", r.getOverview)
}

// listLibraries 获取所有词库摘要列表
func (r *Router) listLibraries(c *gin.Context) {
	// TODO: 调用 librarySvc.ListLibraries()，返回词库摘要列表
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// getLibrary 获取词库详情
func (r *Router) getLibrary(c *gin.Context) {
	// TODO: 从路径参数获取 id，调用 librarySvc.GetLibrary(id)
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// reloadLibraries 重新加载词库（不重启服务）
func (r *Router) reloadLibraries(c *gin.Context) {
	// TODO: 调用 librarySvc.ReloadLibraries()
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// initPlan 初始化14天学习计划
func (r *Router) initPlan(c *gin.Context) {
	// TODO: 从请求体解析 libraryID，调用 planSvc.InitPlan(libraryID)
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// getTodayPlan 获取今日学习计划
func (r *Router) getTodayPlan(c *gin.Context) {
	// TODO: 调用 planSvc.GetTodayPlan()，返回今日新词+复习词
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// getPlanOverview 获取14天计划总览
func (r *Router) getPlanOverview(c *gin.Context) {
	// TODO: 调用 planSvc.GetOverview()，返回整体计划
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// browse 浏览模式：获取今日单词（按分类分组）
func (r *Router) browse(c *gin.Context) {
	// TODO: 调用 studySvc.Browse()，返回分类分组的单词列表
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// markBrowse 浏览标记（认识/不认识）
func (r *Router) markBrowse(c *gin.Context) {
	// TODO: 从请求体解析 wordID 和 known，调用 studySvc.MarkBrowse()
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// generateQuiz 生成测试题目
func (r *Router) generateQuiz(c *gin.Context) {
	// TODO: 从查询参数获取 count，调用 studySvc.GenerateQuiz(count)
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// submitQuizAnswer 提交测试答案
func (r *Router) submitQuizAnswer(c *gin.Context) {
	// TODO: 从请求体解析 wordID 和 answer，调用 studySvc.SubmitQuizAnswer()
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// getDictationWord 获取下一个默写词
func (r *Router) getDictationWord(c *gin.Context) {
	// TODO: 调用 studySvc.GetDictationWord()，返回默写题目
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// submitDictation 提交默写答案
func (r *Router) submitDictation(c *gin.Context) {
	// TODO: 从请求体解析 wordID 和 input，调用 studySvc.SubmitDictation()
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// calcScore 计算今日得分（不打卡）
func (r *Router) calcScore(c *gin.Context) {
	// TODO: 调用 checkInSvc.CalcDailyScore()，返回各模式得分
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// doCheckIn 执行打卡
func (r *Router) doCheckIn(c *gin.Context) {
	// TODO: 从请求体解析 studyDurationSec，调用 checkInSvc.DoCheckIn()
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// getCheckInStats 获取打卡统计
func (r *Router) getCheckInStats(c *gin.Context) {
	// TODO: 调用 checkInSvc.GetStats()，返回打卡统计汇总
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}

// getOverview 获取总览统计数据
func (r *Router) getOverview(c *gin.Context) {
	// TODO: 调用 statsSvc.GetOverview()，返回总览统计
	c.JSON(http.StatusOK, gin.H{"message": "TODO"})
}
