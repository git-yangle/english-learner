package handler

import (
	"embed"
	"io/fs"
	"net/http"
	"strconv"

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
	// 检查是否有计划（前端初始化判断用）
	api.GET("/plan/status", r.getPlanStatus)

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
	list, err := r.librarySvc.ListLibraries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// getLibrary 获取词库详情（含单词列表）
func (r *Router) getLibrary(c *gin.Context) {
	id := c.Param("id")
	library, err := r.librarySvc.GetLibrary(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": library})
}

// reloadLibraries 重新加载词库（不重启服务，热更新）
func (r *Router) reloadLibraries(c *gin.Context) {
	if err := r.librarySvc.ReloadLibraries(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil})
}

// initPlanRequest 初始化计划请求体
type initPlanRequest struct {
	LibraryID string `json:"library_id" binding:"required"`
}

// initPlan 初始化14天学习计划
func (r *Router) initPlan(c *gin.Context) {
	var req initPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	plan, err := r.planSvc.InitPlan(req.LibraryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": plan})
}

// getTodayPlan 获取今日学习计划（含新词+复习词）
func (r *Router) getTodayPlan(c *gin.Context) {
	plan, err := r.planSvc.GetTodayPlan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": plan})
}

// getPlanOverview 获取14天计划总览
func (r *Router) getPlanOverview(c *gin.Context) {
	overview, err := r.planSvc.GetOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": overview})
}

// getPlanStatus 检查是否已有学习计划（前端初始化判断用）
func (r *Router) getPlanStatus(c *gin.Context) {
	has, err := r.planSvc.HasPlan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"has_plan": has}})
}

// browse 浏览模式：获取今日单词（按分类分组）
func (r *Router) browse(c *gin.Context) {
	grouped, err := r.studySvc.Browse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": grouped})
}

// markBrowseRequest 浏览标记请求体
type markBrowseRequest struct {
	WordID string `json:"word_id" binding:"required"`
	Known  bool   `json:"known"`
}

// markBrowse 浏览标记（认识/不认识）
func (r *Router) markBrowse(c *gin.Context) {
	var req markBrowseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := r.studySvc.MarkBrowse(req.WordID, req.Known); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nil})
}

// generateQuiz 生成测试题目（查询参数 ?count=10，默认10题）
func (r *Router) generateQuiz(c *gin.Context) {
	countStr := c.DefaultQuery("count", "10")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "count 参数无效，需为正整数"})
		return
	}
	questions, err := r.studySvc.GenerateQuiz(count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": questions})
}

// submitQuizAnswerRequest 提交测试答案请求体
type submitQuizAnswerRequest struct {
	WordID string `json:"word_id" binding:"required"`
	Answer string `json:"answer" binding:"required"`
}

// submitQuizAnswer 提交测试答案，返回判题结果
func (r *Router) submitQuizAnswer(c *gin.Context) {
	var req submitQuizAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	result, err := r.studySvc.SubmitQuizAnswer(req.WordID, req.Answer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// getDictationWord 获取下一个默写题（优先返回错误率高的词）
func (r *Router) getDictationWord(c *gin.Context) {
	question, err := r.studySvc.GetDictationWord()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": question})
}

// submitDictationRequest 提交默写答案请求体
type submitDictationRequest struct {
	WordID string `json:"word_id" binding:"required"`
	Input  string `json:"input"`
}

// submitDictation 提交默写答案，返回判题结果
func (r *Router) submitDictation(c *gin.Context) {
	var req submitDictationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	result, err := r.studySvc.SubmitDictation(req.WordID, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

// calcScore 计算今日得分（不打卡，仅统计当前学习情况）
func (r *Router) calcScore(c *gin.Context) {
	score, err := r.checkInSvc.CalcDailyScore()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": score})
}

// doCheckInRequest 执行打卡请求体（学习时长可选）
type doCheckInRequest struct {
	StudyDurationSec int `json:"study_duration_sec"`
}

// doCheckIn 执行打卡（计算得分+保存+触发次日复习词生成）
func (r *Router) doCheckIn(c *gin.Context) {
	var req doCheckInRequest
	// ShouldBindJSON 允许空 body，duration 为可选字段
	_ = c.ShouldBindJSON(&req)

	checkIn, err := r.checkInSvc.DoCheckIn(req.StudyDurationSec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": checkIn})
}

// getCheckInStats 获取打卡统计（日历、连续天数等）
func (r *Router) getCheckInStats(c *gin.Context) {
	stats, err := r.checkInSvc.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": stats})
}

// getOverview 获取总览统计数据（打卡日历、得分趋势、分类进度等）
func (r *Router) getOverview(c *gin.Context) {
	overview, err := r.statsSvc.GetOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": overview})
}
