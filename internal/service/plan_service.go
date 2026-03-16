package service

import (
	"english-learner/internal/domain"
	"english-learner/internal/repository"
)

// PlanService 学习计划服务接口
type PlanService interface {
	// InitPlan 初始化14天学习计划
	InitPlan(libraryID string) (*domain.StudyPlan, error)
	// GetTodayPlan 获取今日计划（含新词+复习词）
	GetTodayPlan() (*TodayPlan, error)
	// GetOverview 获取14天总览
	GetOverview() (*domain.StudyPlan, error)
	// HasPlan 检查是否已有计划
	HasPlan() (bool, error)
}

// TodayPlan 今日学习计划（含完整单词信息）
type TodayPlan struct {
	Day         int            `json:"day"`
	Date        string         `json:"date"`
	Status      string         `json:"status"`
	NewWords    []*domain.Word `json:"new_words"`
	ReviewWords []*domain.Word `json:"review_words"`
}

type planService struct {
	planRepo    repository.PlanRepository
	libraryRepo repository.LibraryRepository
	reviewSvc   ReviewService
}

// NewPlanService 创建学习计划服务实例
func NewPlanService(planRepo repository.PlanRepository, libraryRepo repository.LibraryRepository, reviewSvc ReviewService) PlanService {
	return &planService{
		planRepo:    planRepo,
		libraryRepo: libraryRepo,
		reviewSvc:   reviewSvc,
	}
}

// InitPlan 初始化14天学习计划
func (s *planService) InitPlan(libraryID string) (*domain.StudyPlan, error) {
	// TODO:
	// 1. 调用 libraryRepo.GetByID(libraryID) 获取词库
	// 2. 打乱单词顺序，平均分配到14天（每天约15词）
	// 3. 生成 DailyPlan[1..14]，每条含 Day、Date、WordIDs、Status=pending
	// 4. 调用 planRepo.Save() 写入文件
	// 5. 返回生成的 StudyPlan
	return nil, nil
}

// GetTodayPlan 获取今日计划（含新词+复习词完整信息）
func (s *planService) GetTodayPlan() (*TodayPlan, error) {
	// TODO:
	// 1. 调用 planRepo.Get() 获取计划，找到日期匹配今天的 DailyPlan
	// 2. 调用 reviewSvc.GetReviewWords(today) 获取复习词列表
	// 3. 根据 WordIDs 从词库中获取完整单词信息
	// 4. 返回 TodayPlan（含新词列表和复习词列表）
	return nil, nil
}

// GetOverview 获取14天总览
func (s *planService) GetOverview() (*domain.StudyPlan, error) {
	// TODO: 调用 planRepo.Get() 返回完整计划
	return nil, nil
}

// HasPlan 检查是否已有计划
func (s *planService) HasPlan() (bool, error) {
	// TODO: 调用 planRepo.Get()，不为 nil 则返回 true
	return false, nil
}
